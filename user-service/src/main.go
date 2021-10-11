package main

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	pb "github.com/open-exam/open-exam-backend/user-service/grpc-user-service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"net"
	"os"
	"time"
)

var db *sql.DB
var mode = "prod"

func main() {

	if len(os.Args) > 1 && len(os.Args[1]) > 0 {
		mode = os.Args[1]
	}

	if err := godotenv.Load("." + mode + ".env"); err != nil {
		log.Fatalf("Error loading .env file")
	}

	var err error

	var (
		dbUser = os.Getenv("db_user")
		dbPasswd = os.Getenv("db_pass")
		dbHost = os.Getenv("db_host")
		dbPort = os.Getenv("db_port")
		listenAddr = os.Getenv("listen_addr")
	)

	log.Println("user-service starting...")

	dbAddr := net.JoinHostPort(dbHost, dbPort)
	dbConfig := mysql.Config{
		User:      dbUser,
		Passwd:    dbPasswd,
		Net:       "tcp",
		Addr:      dbAddr,
		DBName:    "open_exam",
	}

	for {
		db, err = sql.Open("mysql", dbConfig.FormatDSN())
		if err != nil {
			goto dbError
		}
		err = db.Ping()
		if err != nil {
			goto dbError
		}
		break
		dbError:
			log.Println(err)
			log.Println("error connecting to the database, retrying in 5 secs...")
			time.Sleep(5 * time.Second)
	}

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s, _ := NewServer()
	grpcServer := grpc.NewServer()

	hs := health.NewServer()
	hs.SetServingStatus("grpc.health.v1.user-service", 1)
	healthpb.RegisterHealthServer(grpcServer, hs)

	pb.RegisterUserServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to start grpc server: %v", err)
	}
}
