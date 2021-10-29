package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	rbacPb "github.com/open-exam/open-exam-backend/rbac-service/grpc-rbac-service"
	"github.com/open-exam/open-exam-backend/shared"
	userPb "github.com/open-exam/open-exam-backend/user-db-service/grpc-user-db-service"
)

var (
	errTypeNotFound = errors.New("json: cannot unmarshal type")
)

type TypeSwitch struct {
	Type int32
}

type CreateUser struct {
	TypeSwitch
	CommonFields struct {
		UserId string `json:"user_id" binding:"required"`
		RunScope uint64 `json:"run_scope" binding:"required"`
		Email string `json:"email" binding:"required"`
		Name string `json:"name" binding:"required"`
	}
	*CreateStudent
	*CreateStandardUser
}

type CreateStudent struct {
	TeamId uint64 `json:"team_id"`
}

type CreateStandardUser struct {
	Scope uint64 `json:"scope"`
	ScopeType uint32 `json:"scope_type"`
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
	//router.POST("/generate", generateUsers)
	//router.GET("/", getUserProfile)
	//router.PUT("/", updateUser)
	//router.DELETE("/", deleteUser)
}

func createUser(ctx *gin.Context) {
	var req CreateUser

	if err := ctx.BindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": err.Error(),
		})
		return
	}

	conn, err := shared.GetGrpcConn("rbac-service:" + rbacService)
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.Errors.ServiceConnection)
		return
	}

	rbacClient := rbacPb.NewRbacServiceClient(conn)

	res, err := rbacClient.CanPerformOperation(context.Background(), &rbacPb.CanPerformOperationRequest{
		UserId: req.CommonFields.UserId,
		Scope: req.CommonFields.RunScope,
		Resource: "users",
		Operation: []string {"CREATE"},
	})

	if err != nil || !res.Status {
		ctx.AbortWithStatusJSON(403, gin.H {
			"error": "Invalid scope",
		})
	}

	userClient := userPb.NewUserServiceClient(conn)

	user, err := userClient.CreateUser(context.Background(), &userPb.User {
		Email: req.CommonFields.Email,
		Name:  req.CommonFields.Name,
		Type: uint32(req.Type),
	})
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.Errors.UnknownError)
		return
	}

	userData := &userPb.AddUser {
		Id: user.Id,
	}

	if req.Type == 1 {
		userData.Scope = req.CreateStudent.TeamId
		userData.ScopeType = 4
	} else {
		userData.Scope = req.CreateStandardUser.Scope
		userData.ScopeType = req.CreateStandardUser.ScopeType
	}

	addRes, err := userClient.AddUserToScope(context.Background(), userData)

	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.Errors.UnknownError)
		return
	}

	if !addRes.Status {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": addRes.Message,
		})
		return
	}

	ctx.Status(200)
}