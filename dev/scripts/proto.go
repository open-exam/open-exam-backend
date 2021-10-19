package main

import (
	"fmt"
	"os/exec"
)

var files = []string {
	"user-db-service",
	"exam-db-service",
	"relation-service",
}

func main() {
	generate := func(path string, extraArgs ...string) {
		args := []string {"--go_out=.", "--go_opt=paths=source_relative", "--go-grpc_out=."}
		if len(extraArgs) > 0 {
			args = append(args, extraArgs...)
		}
		args = append(args, "--go-grpc_opt=paths=source_relative", path)

		cmd := exec.Command("protoc", args...)
		res, err := cmd.Output()
		if err != nil {
			fmt.Println(path, err.Error(), string(res))
		}
	}

	for _, file := range files {
		generate(file + "/grpc-" + file + "/" + file + ".proto", "-I=grpc-shared", "--proto_path=.")
	}


	generate("grpc-shared/shared.proto")
}