package main

import (
	"crypto/rsa"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/open-exam/open-exam-backend/shared"
	"github.com/open-exam/open-exam-backend/util"
	"log"
	"os"
)

var (
	mode = "prod"
	jwtPublicKey *rsa.PublicKey
	rbacService string
	userService string
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
}
