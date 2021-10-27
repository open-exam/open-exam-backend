package main

import (
	"github.com/gin-gonic/gin"
	"github.com/open-exam/open-exam-backend/shared"
	"log"
	"os"
)

var (
	mode = "prod"
)

// @title open-exam public API
// @version 0.1
// @description The open-exam publicly exposed API.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url https://github.com/open-exam/open-exam-backend/blob/master/LICENSE

// @host localhost:8080
// @BasePath /
// @schemes http

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

	if err := router.Run(listenAddr); err != nil {
		log.Fatalf("failed to start oauth2 server: %v", err)
	}
}

func validateOptions() {

}
