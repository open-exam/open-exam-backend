package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	rbacPb "github.com/open-exam/open-exam-backend/rbac-service/grpc-rbac-service"
	"github.com/open-exam/open-exam-backend/shared"
	userPb "github.com/open-exam/open-exam-backend/user-db-service/grpc-user-db-service"
	"github.com/open-exam/open-exam-backend/util"
)

var (
	errTypeNotFound = errors.New("json: cannot unmarshal type")
	emailRegexp     = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type TypeSwitch struct {
	Type int32
}

type CreateUser struct {
	TypeSwitch
	CommonFields struct {
		UserId   string `json:"user_id" binding:"required"`
		RunScope uint64 `json:"run_scope" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Name     string `json:"name" binding:"required"`
	}
	*CreateStudent
	*CreateStandardUser
}

type CreateStudent struct {
	TeamId uint64 `json:"team_id"`
}

type CreateStandardUser struct {
	Scope     uint64 `json:"scope"`
	ScopeType uint32 `json:"scope_type"`
}

type ErrorList struct {
	item  string
	error string
}

type ParsedUser struct {
	Name      string
	Email     string
	Type      uint32
	Scope     uint64
	ScopeType uint32
	Count     uint32
}

func (res *CreateUser) UnmarshalJSON(data []byte) error {
	res.Type = -1
	if err := json.Unmarshal(data, &res.TypeSwitch); err != nil || res.Type == -1 {
		if res.Type == -1 {
			return errTypeNotFound
		}

		return err
	}

	if err := json.Unmarshal(data, &res.CommonFields); err != nil {
		return err
	}

	if res.Type == 0 {
		res.CreateStandardUser = &CreateStandardUser{}
		return json.Unmarshal(data, res.CreateStandardUser)
	}

	res.CreateStudent = &CreateStudent{}
	return json.Unmarshal(data, res.CreateStudent)
}

func InitUsers(router *gin.RouterGroup) {
	router.Use(shared.JwtMiddleware(jwtPublicKey))

	router.POST("/", createUser)
	router.POST("/generate", generateUsers)
	router.GET("/", getUser)
	//router.PUT("/", updateUser)
	//router.DELETE("/", deleteUser)
}

func createUser(ctx *gin.Context) {
	var req CreateUser

	if err := ctx.BindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	conn, err := shared.GetGrpcConn("rbac-service:" + rbacService)
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.ServiceConnection)
		return
	}

	rbacClient := rbacPb.NewRbacServiceClient(conn)

	res, err := rbacClient.CanPerformOperation(context.Background(), &rbacPb.CanPerformOperationRequest{
		UserId:    req.CommonFields.UserId,
		Scope:     req.CommonFields.RunScope,
		Resource:  "users",
		Operation: []string{"CREATE"},
	})

	if err != nil || !res.Status {
		ctx.AbortWithStatusJSON(403, gin.H{
			"error": "Invalid scope",
		})
	}

	userClient := userPb.NewUserServiceClient(conn)

	user, err := userClient.CreateUser(context.Background())
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.UnknownError)
		return
	}

	err = user.Send(&userPb.User{
		Email: req.CommonFields.Email,
		Name:  req.CommonFields.Name,
		Type:  uint32(req.Type),
	})
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.UnknownError)
		return
	}

	responseUser, err := user.Recv()
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.UnknownError)
		return
	}

	if len(responseUser.Error) != 0 {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": responseUser.Error,
		})
		return
	}

	userData := &userPb.AddUser{
		Id: responseUser.Id,
	}

	if req.Type == 1 {
		userData.Scope = req.CreateStudent.TeamId
		userData.ScopeType = 4
	} else {
		userData.Scope = req.CreateStandardUser.Scope
		userData.ScopeType = req.CreateStandardUser.ScopeType
	}

	addScopeStream, err := userClient.AddUserToScope(context.Background())
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.UnknownError)
		return
	}

	err = addScopeStream.Send(userData)
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.UnknownError)
		return
	}

	addRes, err := addScopeStream.Recv()
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.UnknownError)
		return
	}

	if !addRes.Status {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": addRes.Message,
		})
		return
	}

	ctx.Status(200)
}

func generateUsers(ctx *gin.Context) {
	if ctx.Request.Header.Get("content-type") != "application/csv" {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "invalid content-type",
		})
		return
	}

	buf, err := ctx.GetRawData()
	if err != nil {
		ctx.AbortWithStatusJSON(400, shared.GinErrors.UnknownError)
	}

	records, err := util.ReadCSV(buf)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "invalid csv file",
		})
		return
	}

	if len(records) == 0 {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "no records",
		})
		return
	}

	if len(records[0]) != 5 {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "wrong number of records",
		})
		return
	}

	if strings.ToLower(records[0][0]) == "name" {
		records = records[1:]
	} else if strings.ToLower(records[0][0]) == "email" {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "csv not in correct format",
		})
		return
	}

	errorList := make([]ErrorList, 0)
	newData := make([]ParsedUser, 0)

	count := 0
	removeSlice := func(slice [][]string, index int) [][]string {
		return append(slice[:index], slice[index+1:]...)
	}

	for i, row := range records {
		if len(row) == 2 {
			if len(row[0]) == 0 {
				errorList = append(errorList, ErrorList{item: row[0], error: "empty name"})
				removeSlice(records, i-count)
				count++
				continue
			}

			if !emailRegexp.MatchString(row[1]) {
				errorList = append(errorList, ErrorList{item: row[1], error: "invalid email address"})
				removeSlice(records, i-count)
				count++
				continue
			}

			scope, err := strconv.ParseUint(row[2], 10, 64)
			if err != nil {
				errorList = append(errorList, ErrorList{item: row[2], error: "invalid scope"})
				removeSlice(records, i-count)
				count++
				continue
			}

			scopeType, err := strconv.ParseUint(row[3], 10, 32)
			if err != nil {
				errorList = append(errorList, ErrorList{item: row[3], error: "invalid scope type"})
				removeSlice(records, i-count)
				count++
				continue
			}

			userType := uint32(0)
			if row[4] == "1" {
				userType = 1
			}

			newData = append(newData, ParsedUser{
				Name:      row[0],
				Email:     row[1],
				Type:      userType,
				Scope:     scope,
				ScopeType: uint32(scopeType),
			})
		} else {
			errorList = append(errorList, ErrorList{item: row[1], error: "invalid number of items"})
			removeSlice(records, i-count)
			count++
		}
	}

	conn, err := shared.GetGrpcConn("rbac-service:" + rbacService)
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.ServiceConnection)
		return
	}

	userClient := userPb.NewUserServiceClient(conn)
	createUserStream, err := userClient.CreateUser(context.Background())
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.ServiceConnection)
		return
	}

	go func() {
		removeItem := func(slice []ParsedUser, index int) []ParsedUser {
			return append(slice[:index], slice[index+1:]...)
		}

		for {
			res, err := createUserStream.Recv()

			if err == io.EOF {
				break
			}

			if err != nil {
				continue
			}

			email, idx := findEmail(&newData, res.Count)
			if len(res.Error) > 0 {
				errorList = append(errorList, ErrorList{
					item:  email,
					error: err.Error(),
				})
			}

			removeItem(newData, idx)
		}

		if len(newData) > 0 {
			for _, item := range newData {
				errorList = append(errorList, ErrorList{
					item:  item.Email,
					error: "unknown error",
				})
			}
		}

		ctx.JSON(200, errorList)
	}()

	count = 0
	for i, data := range newData {
		err := createUserStream.Send(&userPb.User{
			Name:  data.Name,
			Email: data.Email,
			Type:  data.Type,
			Count: uint32(i - count),
		})

		if err != nil {
			errorList = append(errorList, ErrorList{item: data.Email, error: err.Error()})
			removeSlice(records, i-count)
			count++
			continue
		}

		newData[i-count-1].Count = uint32(i - count - 1)
	}
}

func getUser(ctx *gin.Context) {

}

func findEmail(data *[]ParsedUser, idx uint32) (string, int) {
	for i, e := range *data {
		if e.Count == idx {
			return e.Email, i
		}
	}
	return "", 0
}
