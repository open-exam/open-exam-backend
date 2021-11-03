package main

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	rbacPb "github.com/open-exam/open-exam-backend/rbac-service/grpc-rbac-service"
	"github.com/open-exam/open-exam-backend/shared"
)

func InitExams(router *gin.RouterGroup) {
	router.Use(shared.JwtMiddleware(jwtPublicKey, mode))

	router.POST("/", createExam)
}

func createExam(ctx *gin.Context) {
	user, _ := ctx.Get("user")
	tempScope := ctx.Query("scope")
	if len(tempScope) == 0 {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": "scope is empty",
		})
		return
	}

	scope, err := strconv.ParseUint(tempScope, 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": "invalid scope",
		})
		return
	}

	conn, err := shared.GetGrpcConn(rbacService)
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.ServiceConnection)
		return
	}

	rbacClient := rbacPb.NewRbacServiceClient(conn)
	res, err := rbacClient.CanPerformOperation(context.Background(), &rbacPb.CanPerformOperationRequest {
		UserId: user.(string),
		Resource: "EXAMS",
		Operation: []string{"CREATE"},
		Scope: scope,
	})

	if !res.Status {
		ctx.AbortWithStatusJSON(403, gin.H {
			"error": "you do not have permission to perform this operation",
		})
		return
	}

	buf, err := ctx.GetRawData()
	if err != nil {
		ctx.AbortWithStatusJSON(400, shared.GinErrors.UnknownError)
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
}