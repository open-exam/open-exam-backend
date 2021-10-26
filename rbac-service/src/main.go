package main

import (
	"database/sql"
	"os"

	pb "github.com/open-exam/open-exam-backend/rbac-service/grpc-rbac-service"
	"github.com/open-exam/open-exam-backend/shared"
	"google.golang.org/grpc"
)

var (
	db *sql.DB
	mode = "prod"
	relationService string
)

func main() {

	shared.SetEnv(&mode)

	validateOptions()

	shared.DefaultGrpcServer(db, func(server *grpc.Server) {
		s, _ := NewServer()
		pb.RegisterRbacServiceServer(server, s)
	})
}


func validateOptions() {
	relationService = os.Getenv("relation_service")
}