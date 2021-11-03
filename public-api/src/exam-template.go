package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	examPb "github.com/open-exam/open-exam-backend/exam-db-service/grpc-exam-db-service"
	"github.com/open-exam/open-exam-backend/shared"
	"github.com/open-exam/open-exam-backend/util"
)



func InitExamTemplates(router *gin.RouterGroup) {
	router.Use(shared.JwtMiddleware(jwtPublicKey, mode))

	router.POST("/", createExamTemplate)
}

func createExamTemplate(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": err.Error(),
		})
		return
	}

	orgIds := form.Value["org_id"]
	if len(orgIds) == 0 || len(orgIds[0]) == 0 {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "no org_id",
		})
		return
	}

	_, err = strconv.ParseUint(orgIds[0], 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "invalid organization id",
		})
		return
	}

	names := form.Value["name"]
	if len(names) == 0 || len(names[0]) == 0 {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "no name",
		})
		return
	}
	
	scopes := form.Value["scopes"]
	if len(scopes) == 0 {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "no scopes",
		})
		return
	}

	parsedScopes := make([]Scope, 0)
	err = json.Unmarshal([]byte(scopes[0]), &parsedScopes)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "invalid scopes",
		})
		return
	}
	
	files := form.File["template"]
	if len(files) == 0 {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": "No template file provided",
		})
		return
	}
	
	buf := make([]byte, files[0].Size)
	file, err := files[0].Open()
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": shared.GinErrors.UnknownError,
		})
		return
	}

	n, err := file.Read(buf)
	if err != nil || n == 0 {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": shared.GinErrors.UnknownError,
		})
		return
	}

	examConfig := &shared.Exam{}
	err = json.Unmarshal(buf, examConfig)
	if err != nil {
		ctx.AbortWithStatusJSON(400, shared.GinErrors.JsonParseError)
		return
	}

	if err := examConfig.Validate(); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": err.Error(),
		})
		return
	}

	ct, body, err := util.CreateMultipartForm(map[string]string {
		"template": string(buf),
		"org_id": form.Value["org_id"][0],
	})
	conn, err := shared.GetGrpcConn(examService)
	if err != nil {
		ctx.AbortWithStatusJSON(400, shared.GinErrors.ServiceConnection)
		return
	}

	examClient := examPb.NewExamTemplateClient(conn)
	
	examScopes := make([]*examPb.Scope, 0)

	for _, sc := range parsedScopes {
		examScopes = append(examScopes, &examPb.Scope {
			Scope: sc.Scope,
			ScopeType: sc.ScopeType,
		})
	}

	res, err := examClient.CreateTemplate(ctx, &examPb.CreateExamTemplateRequest{
		Name: names[0],
		Scopes: examScopes,
	})
	if err != nil {
		ctx.AbortWithStatusJSON(400, err)
		return
	}

	resp, err := http.Post(fsService + "/exam-template", ct, body)
	if err != nil || resp.StatusCode >= 400 {
		ctx.AbortWithStatusJSON(400, shared.GinErrors.UnknownError)
		return
	}

	ctx.JSON(200, gin.H {
		"id": res.IdString,
	})
}