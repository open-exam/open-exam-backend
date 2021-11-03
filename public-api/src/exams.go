package main

import (
	"context"

	"github.com/gin-gonic/gin"
	examPb "github.com/open-exam/open-exam-backend/exam-db-service/grpc-exam-db-service"
	sharedPb "github.com/open-exam/open-exam-backend/grpc-shared"
	relationPb "github.com/open-exam/open-exam-backend/relation-service/grpc-relation-service"
	"github.com/open-exam/open-exam-backend/shared"
)

type Exam struct {
	Name string `json:"name" binding:"required"`
	StartTime uint64 `json:"start_time" binding:"required"`
	EndTime uint64 `json:"end_time" binding:"required"`
	Duration uint32 `json:"duration" binding:"required"`
	Template string `json:"template" binding:"required"`
	Organization uint64 `json:"organization" binding:"required"`
	Scopes []Scope `json:"scopes" binding:"required"`
}

func InitExams(router *gin.RouterGroup) {
	router.Use(shared.JwtMiddleware(jwtPublicKey, mode))

	router.POST("/", createExam)
}

func createExam(ctx *gin.Context) {
	exam := Exam{}
	if ctx.BindJSON(&exam) != nil {
		ctx.AbortWithStatusJSON(400, shared.GinErrors.JsonParseError)
		return
	}

	conn, err := shared.GetGrpcConn(examService)
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.ServiceConnection)
		return
	}

	examTemplateClient := examPb.NewExamTemplateClient(conn)

	userId, _ := ctx.Get("user")
	res, err := examTemplateClient.ValidateTemplate(context.Background(), &sharedPb.StandardIdRequest {
		IdString: exam.Template,
	})
	if err != nil {
		ctx.AbortWithStatusJSON(400, err)
		return
	}

	if !res.Status {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": res.Message,
		})
		return
	}

	relationClient := relationPb.NewRelationServiceClient(conn)

	res, err = relationClient.CanAccessTemplate(context.Background(), &relationPb.CanAccessTemplateRequest {
		UserId: userId.(string),
		TemplateId: exam.Template,
	})
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.UnknownError)
		return
	}

	if !res.Status {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": res.Message,
		})
		return
	}

	examClient := examPb.NewExamServiceClient(conn)

	scopes := make([]*examPb.Scope, len(exam.Scopes))
	for i, sc := range exam.Scopes {
		scopes[i] = &examPb.Scope {
			Scope: sc.Scope,
			ScopeType: sc.ScopeType,
		}
	}

	createRes, err := examClient.CreateExam(context.Background(), &examPb.CreateExamRequest {
		Name: exam.Name,
		StartTime: exam.StartTime,
		EndTime: exam.EndTime,
		Duration: exam.Duration,
		Template: exam.Template,
		Organization: exam.Organization,
		CreatedBy: userId.(string),
		Scopes: scopes,
	})
	if err != nil {
		ctx.AbortWithStatusJSON(400, err)
		return
	}

	// TODO: run validateTemplate from exam-db-service to validate all question Ids pool Ids access scopes, etc.

	ctx.JSON(200, gin.H {
		"id": createRes.IdString,
	})
}