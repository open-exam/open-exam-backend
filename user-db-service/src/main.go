package main

import (
	"database/sql"
	"github.com/open-exam/open-exam-backend/shared"
	pb "github.com/open-exam/open-exam-backend/user-db-service/grpc-user-db-service"
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
