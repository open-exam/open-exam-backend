package main

import (
	"database/sql"
	"github.com/open-exam/open-exam-backend/shared"
	pb "github.com/open-exam/open-exam-backend/user-service/grpc-user-service"
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
		pb.RegisterUserServiceServer(server, s)
	})
}
