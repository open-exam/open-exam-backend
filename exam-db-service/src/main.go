package main

import (
	"database/sql"

	pb "github.com/open-exam/open-exam-backend/exam-db-service/grpc-exam-db-service"
	"github.com/open-exam/open-exam-backend/shared"
	"google.golang.org/grpc"
)

var (
	db   *sql.DB
	mode = "prod"
)

func main() {
	shared.SetEnv(&mode)

	shared.DefaultGrpcServer(func(server *grpc.Server) {
		db = shared.Db
		sExamClient, _ := NewExamClientAccessServer()
		sExamService, _ := NewExamServiceServer()
		sExamTemplate, _ := NewExamTemplateServer()

		pb.RegisterExamClientAccessServer(server, sExamClient)
		pb.RegisterExamServiceServer(server, sExamService)
		pb.RegisterExamTemplateServer(server, sExamTemplate)
	})
}
