package shared

import (
	"crypto/rsa"
	"database/sql"
	"errors"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type GinErrorList struct {
	ServiceConnection gin.H
	UnknownError gin.H
	JsonParseError gin.H
}

type StandardErrorList struct {
	ServiceConnection error
	UnknownError error
}

var (
	errUnexpectedSigningMethod = errors.New("unknown signing method")
	GinErrors                  = GinErrorList{
		ServiceConnection: gin.H {
			"error": "could not connect to internal service",
		},
		UnknownError: gin.H {
			"error": "an unknown error occurred",
		},
		JsonParseError: gin.H {
			"error": "could not parse JSON",
		},
	}
	Errors = StandardErrorList {
		ServiceConnection: errors.New("could not connect to service"),
		UnknownError: errors.New("an unknown error occurred"),
	}
)

func SetEnv(mode *string) {
	devMode := flag.Bool("dev", false, "Run in dev mode")
	envFile := flag.String("env", "", ".env file path")
	flag.Parse()

	if *devMode {
		*mode = "dev"
	}

	if err := godotenv.Load(func() string {
		if len(*envFile) > 0 {
			return *envFile
		}

		return "." + *mode + ".env"
	}()); err != nil {
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

	log.Println("user-db-service starting...")

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
	hs.SetServingStatus("grpc.health.v1.user-db-service", 1)
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

func JwtMiddleware(publicKey *rsa.PublicKey, mode string) gin.HandlerFunc {
	return func (ctx *gin.Context) {
		if mode == "dev" {
			ctx.Next()
			return
		}
		header := ctx.Request.Header.Get("Authorization")

		if len(header) == 0 {
			ctx.AbortWithStatus(401)
			return
		}

		splitHeader := strings.Split(header, "Bearer")
		if len(splitHeader) != 2 || (len(splitHeader) == 2 && len(strings.TrimSpace(splitHeader[1])) == 0) {
			ctx.AbortWithStatus(401)
			return
		}

		tok, err := jwt.Parse(strings.TrimSpace(splitHeader[1]),
			func (jwtToken *jwt.Token) (interface{}, error) {
				if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, errUnexpectedSigningMethod
				}
				return publicKey, nil
			},
		)

		if err != nil {
			ctx.AbortWithStatusJSON(401, err)
			return
		}

		claims, ok := tok.Claims.(jwt.MapClaims)
		if !ok || !tok.Valid {
			ctx.AbortWithStatus(403)
			return
		}

		ctx.Set("user", claims["user"])
		ctx.Set("data", claims["data"])
		ctx.Next()
	}
}