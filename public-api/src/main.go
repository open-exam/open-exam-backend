package main

import (
	"crypto/rsa"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/open-exam/open-exam-backend/shared"
	"github.com/open-exam/open-exam-backend/util"
)

var (
	mode = "prod"
	jwtPublicKey *rsa.PublicKey
	rbacService string
	userService string
	fsService string
	examService string
)

func main() {
	shared.SetEnv(&mode)
	gin.SetMode(gin.DebugMode)

	if mode == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	validateOptions()

	listenAddr := os.Getenv("listen_addr")

	router := gin.New()
	router.Use(gin.Recovery())
	if mode == "dev" {
		router.Use(gin.Logger())
	}

	InitUsers(router.Group("/users"))
	InitExams(router.Group("/exams"))
	InitExamTemplates(router.Group("/exam-template"))
	InitPlugins(router.Group("/plugins"))

	if err := router.Run(listenAddr); err != nil {
		log.Fatalf("failed to start oauth2 server: %v", err)
	}
}

func validateOptions() {
	tempJwtPublicKey, err := util.DecodeBase64([]byte(os.Getenv("jwt_public_key")))
	if err != nil {
		log.Fatalf("invalid jwt_public_key")
	}
	jwtPublicKey, err = jwt.ParseRSAPublicKeyFromPEM(tempJwtPublicKey)
	if err != nil {
		log.Fatalf("invalid jwt_public_key")
	}

	rbacService = os.Getenv("rbac_service")
	userService = os.Getenv("user_service")
	fsService = os.Getenv("fs_service")
	examService = os.Getenv("exam_service")
}
