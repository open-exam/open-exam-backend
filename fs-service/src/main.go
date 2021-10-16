package main

import (
	"github.com/gin-gonic/gin"
	"github.com/open-exam/open-exam-backend/shared"
	"github.com/open-exam/open-exam-backend/util"
	"log"
	"os"
)

var (
	mode = "prod"
	errServiceConnection = gin.H {
		"error": "server_error",
		"error_description": "501; could not connect to internal service",
	}
)

func main() {
	shared.SetEnv(&mode)

	validateOptions()

	listenAddr := os.Getenv("listen_addr")

	router := gin.New()
	router.Use(gin.Recovery())

	InitExamFiles(router.Group("/test-files"))

	if err := router.Run(listenAddr); err != nil {
		log.Fatalf("failed to start oauth2 server: %v", err)
	}
}

func validateOptions() {
	platforms = util.SplitAndParse(os.Getenv("platforms"))

	examClientAccessService = os.Getenv("exam_client_access_service")

	relationService = os.Getenv("relation_service")

	clientName = os.Getenv("client_name")
}