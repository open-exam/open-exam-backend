package main

import (
	"database/sql"
	pb "github.com/open-exam/open-exam-backend/relation-service/grpc-relation-service"
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
		s, _ := NewServer()
		pb.RegisterRelationServiceServer(server, s)
	})
}
