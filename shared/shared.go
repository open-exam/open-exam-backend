package shared

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"net"
	"os"
	"time"
)

func SetEnv(mode *string) {
	if len(os.Args) > 1 && len(os.Args[1]) > 0 {
		*mode = os.Args[1]
	}
	if err := godotenv.Load("." + *mode + ".env"); err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func DefaultGrpcServer (db *sql.DB, registerComponents func(*grpc.Server)) {

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

	grpcServer := grpc.NewServer()

	hs := health.NewServer()
	hs.SetServingStatus("grpc.health.v1.user-service", 1)
	healthpb.RegisterHealthServer(grpcServer, hs)

	registerComponents(grpcServer)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to start grpc server: %v", err)
	}
}

func GetGrpcConn(connectionString string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(connectionString, grpc.WithInsecure())
	if err != nil {
		defer conn.Close()
	}
	return conn, err
}