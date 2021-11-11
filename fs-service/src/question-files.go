package main

import "github.com/gin-gonic/gin"

func InitQuestionFiles(router *gin.RouterGroup) {
	router.POST("/", createQuestionFile)
	router.GET("/", readQuestionFile)
}

func createQuestionFile(ctx *gin.Context) {
	// TODO: read rbac to ensure create rights are present
}

func readQuestionFile(ctx *gin.Context) {

}
