package main

import (
	"github.com/gin-gonic/gin"
	"github.com/open-exam/open-exam-backend/shared"
)

func InitUsers(router *gin.RouterGroup) {
	router.Use(shared.JwtMiddleware())

	router.POST("/", createUser)
	router.POST("/generate", generateUsers)
	router.GET("/", getUserProfile)
	router.PUT("/", updateUser)
	router.DELETE("/", deleteUser)
}