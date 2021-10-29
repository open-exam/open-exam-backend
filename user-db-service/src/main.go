package main

import (
	"database/sql"
	"github.com/open-exam/open-exam-backend/shared"
	pb "github.com/open-exam/open-exam-backend/user-db-service/grpc-user-db-service"
	"google.golang.org/grpc"
	"os"
	"strconv"
)

var (
	db *sql.DB
	mode = "prod"
	standardPasswordSize uint32
)

func main() {
	shared.SetEnv(&mode)

	validateOptions()

	shared.DefaultGrpcServer(db, func(server *grpc.Server) {
		s, _ := NewServer()
		pb.RegisterUserServiceServer(server, s)
	})
}

func validateOptions() {
	tempSize, _ := strconv.ParseUint(os.Getenv("standard_password_size"), 10, 32)
	standardPasswordSize = uint32(tempSize)
}