package main

import (
	"context"
	"os"
	"strconv"
	questionPb "github.com/open-exam/open-exam-backend/question-service/grpc-question-service"
	"github.com/gin-gonic/gin"
	"github.com/open-exam/open-exam-backend/rbac-service/grpc-rbac-service"
	"github.com/open-exam/open-exam-backend/shared"
)

type Question struct {
	pluginId    uint64  `json:"start_time" binding:"required"`
	displayData      string  `json:"end_time" binding:"required"`
	title     string  `json:"duration" binding:"required"`
	files    string  `json:"template" binding:"required"`
}

func InitQuestionFiles(router *gin.RouterGroup) {
	router.Use(shared.JwtMiddleware(jwtPublicKey, mode))
	router.POST("/", shared.RBACMiddleware(rbacService, "QUESTIONS", []string{"CREATE", "UPDATE"}), createQuestionFile)
	router.GET("/", readQuestionFile)
}

func createQuestionFile(ctx *gin.Context) {
	// TODO: read rbac to ensure create rights are present
	question := Question{}
	var (
		question_id = ctx.Query("question_id")
	)
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	files := form.File["question"]
	orgIds := form.Value["org_id"]
	if len(files) == 0 {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": "no question file",
		})
		return
	}
	if len(orgIds) == 0 || len(orgIds[0]) == 0 {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": "no org_id",
		})
		return
	}
	_, err = strconv.ParseUint(orgIds[0], 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": "Invalid Organization ID",
		})
		return
	}
	buf := make([]byte, files[0].Size)
	file, err := files[0].Open()
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": shared.GinErrors.UnknownError,
		})
	}
	n, err := file.Read(buf)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": shared.GinErrors.UnknownError,
		})
	}
	err = os.WriteFile("/app-data/questions/"+orgIds[0]+"/"+files[0].Filename, buf, 0644)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": shared.GinErrors.UnknownError,
		})
		return
	}
	question.files = files[0].Filename
	conn, err := shared.GetGrpcConn(questionService)
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.ServiceConnection)
		return
	}

	questionServiceClient := questionPb.NewQuestionServiceClient(conn)
	res, err = questionServiceClient.updateQuestion(context.Background(), &questionPb.QuestionDetails{
		questionId:     question_id,
		data: question,
	})
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.UnknownError)
		return
	}
	ctx.Status(200)
}

func readQuestionFile(ctx *gin.Context) {
//make a call to rbac only if it is a student
}
