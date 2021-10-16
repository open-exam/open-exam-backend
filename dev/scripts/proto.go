package main

import (
	"fmt"
	"os/exec"
)

var files = []string {
	"user-service",
	"fs-service",
	"exam-client-access-service",
	"relation-service",
}

func main() {
	for _, file := range files {
		cmd := exec.Command("protoc", "--go_out=.", "--go_opt=paths=source_relative", "--go-grpc_out=.", "--go-grpc_opt=paths=source_relative", file + "/grpc-" + file + "/" + file + ".proto")
		_, err := cmd.Output()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}