package shared

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"github.com/mholt/archiver/v3"
	rbacPb "github.com/open-exam/open-exam-backend/rbac-service/grpc-rbac-service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
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
	Db *sql.DB
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

func DefaultGrpcServer(registerComponents func(*grpc.Server)) {
	var err error
	
	var (
		dbUser = os.Getenv("db_user")
		dbPasswd = os.Getenv("db_pass")
		dbHost = os.Getenv("db_host")
		dbPort = os.Getenv("db_port")
		listenAddr = os.Getenv("listen_addr")

	)

	log.Println("service starting...")

	dbAddr := net.JoinHostPort(dbHost, dbPort)
	dbConfig := mysql.Config{
		User:      dbUser,
		Passwd:    dbPasswd,
		Net:       "tcp",
		Addr:      dbAddr,
		DBName:    "open_exam",
	}

	for {
		Db, err = sql.Open("mysql", dbConfig.FormatDSN())
		if err != nil {
			goto dbError
		}
		err = Db.Ping()
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
			ctx.Set("user", UserJWT{UserId: "1", Scope: 34535})
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

func RBACMiddleware(rbacService string, resource string, operations []string) gin.HandlerFunc {
	conn, err := GetGrpcConn(rbacService)
	if err != nil {
		log.Fatalf("could not connect to rbac service: %v", err)
	}

	rbacClient := rbacPb.NewRbacServiceClient(conn)

	return func (ctx *gin.Context) {
		user, exists := ctx.Get("user")
		if !exists {
			ctx.AbortWithStatus(401)
			return
		}

		res, err := rbacClient.CanPerformOperation(context.Background(), &rbacPb.CanPerformOperationRequest{
			UserId: user.(UserJWT).UserId,
			Resource: resource,
			Operation: operations,
			Scope: user.(UserJWT).Scope,
		})

		if err != nil {
			ctx.AbortWithStatusJSON(500, err)
			return
		}

		if !res.Status {
			ctx.AbortWithStatus(403)
			return
		}

		ctx.Next()
	}
}

func GetPluginSources(Id string, name string, uri string, uriType string) (int, error) {
	switch uriType {
		case "git": {
			_, err := git.PlainClone(name, false, &git.CloneOptions{
				URL: uri,
			})
			if err != nil {
				return 400, errors.New("could not clone repository: " + err.Error())
			}
		}
		case "file": {
			resp, err := http.Get(uri)
			if err != nil {
				return 400, errors.New("could not download plugin: " + err.Error())
			}
			defer resp.Body.Close()
	
			if resp.StatusCode != 200 {
				return 400, errors.New("could not download plugin: " + resp.Status)
			}
	
			out, err := os.Create(name + "/" + Id)
			if err != nil {
				return 500, Errors.UnknownError
			}
			defer out.Close()
	
			_, err = io.Copy(out, resp.Body)
			if err != nil {
				return 500, Errors.UnknownError
			}
	
			err = archiver.Unarchive(name + "/" + Id, name)
			if err != nil {
				return 400, errors.New("could not unarchive plugin: " + err.Error())
			}
		}
		default: {
			return 400, errors.New("invalid uri_type: " + uriType)
		}
	}

	return 200, nil
}