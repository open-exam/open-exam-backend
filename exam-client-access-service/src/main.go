package main

import (
	"database/sql"
	pb "github.com/open-exam/open-exam-backend/exam-client-access-service/grpc-exam-client-access-service"
	"github.com/open-exam/open-exam-backend/shared"
	"google.golang.org/grpc"
)

var (
	db *sql.DB
	mode = "prod"
)

func main() {

	shared.SetEnv(&mode)

	shared.DefaultGrpcServer(db, func(server *grpc.Server) {
		s, _ := NewServer()
		pb.RegisterExamClientAccessServer(server, s)
	})
}
