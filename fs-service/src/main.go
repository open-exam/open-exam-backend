package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/open-exam/open-exam-backend/shared"
	"github.com/open-exam/open-exam-backend/util"
)

var (
	mode = "prod"
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

	InitExamFiles(router.Group("/exam-files"))
	InitExamLog(router.Group("/exam-log"))
	InitQuestionFiles(router.Group("/question-files"))
	InitExamTemplate(router.Group("/exam-template"))

	if err := router.Run(listenAddr); err != nil {
		log.Fatalf("failed to start oauth2 server: %v", err)
	}
}

func validateOptions() {
	platforms = util.SplitAndParse(os.Getenv("platforms"))

	examDbService = os.Getenv("exam_db_service")

	relationService = os.Getenv("relation_service")

	clientName = os.Getenv("client_name")
}
