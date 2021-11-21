package main

import (
	"bytes"
	"database/sql"
	"log"
	"os"
	"os/exec"

	pb "github.com/open-exam/open-exam-backend/container-build-service/grpc-container-build-service"
	"github.com/open-exam/open-exam-backend/shared"
	"google.golang.org/grpc"
)

var (
	db               *sql.DB
	mode             = "prod"
	pluginDbService  string
	registryHost     string
	registryPort     string
	registryLogin    string
	registryPassword string
)

func main() {
	shared.SetEnv(&mode)

	validateOptions()

	podmanLogin()

	shared.DefaultGrpcServer(func(server *grpc.Server) {
		db = shared.Db
		s, _ := NewServer()

		pb.RegisterContainerServer(server, s)
	})
}

func validateOptions() {
	pluginDbService = os.Getenv("plugin_db_service")
	registryHost = os.Getenv("registry_host")
	registryPort = os.Getenv("registry_port")
	registryLogin = os.Getenv("registry_login")
	registryPassword = os.Getenv("registry_password")
}

func podmanLogin() {
	cmd := exec.Command("podman", "login", registryHost+":"+registryPort)
	buffer := bytes.Buffer{}
	buffer.Write([]byte(registryLogin + "\n" + registryPassword + "\n"))
	cmd.Stdin = &buffer

	err := cmd.Wait()
	if err != nil {
		log.Fatal("Could not login to registry: ", err)
	}
}
