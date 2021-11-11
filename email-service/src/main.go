package main

import (
	"database/sql"
	"os"
	"strconv"

	pb "github.com/open-exam/open-exam-backend/email-service/grpc-email-service"
	"github.com/open-exam/open-exam-backend/shared"
	"google.golang.org/grpc"
)

var (
	mode string
	db *sql.DB
	userService string
	templateService string
	fsService string
	smtpHost string
	emailUser string
	emailPassword string
	smtpPort int
	rateLimit int
)

func main() {
	shared.SetEnv(&mode)

	validateOptions()

	shared.DefaultGrpcServer(func(server *grpc.Server) {
		db = shared.Db
		s, _ := NewServer()

		pb.RegisterEmailServiceServer(server, s)
	})
}

func validateOptions() {
	userService = os.Getenv("user_service")
	templateService = os.Getenv("template_service")
	fsService = os.Getenv("fs_service")
	smtpHost = os.Getenv("smtp_host")
	temp, _ := strconv.ParseInt(os.Getenv("smtp_port"), 10, 32)
	smtpPort = int(temp)
	emailUser = os.Getenv("email_user")
	emailPassword = os.Getenv("email_password")
	temp, _ = strconv.ParseInt(os.Getenv("rate_limit"), 10, 32)
	rateLimit = int(temp)
}