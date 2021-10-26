package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var files = []string {}


func main() {
	mode := flag.Bool("all", false, "Use all files")
	inFile := flag.String("in", "", "Input folder path")

	flag.Parse()
	if *mode {
		data, err := os.ReadFile(".github/grpc-ci-paths")
		if err != nil {
			log.Fatal(err)
		}

		files = func () []string {
			items := strings.Split(string(data), "\n")
			res := make([]string, 0)

			for _, f := range items {
				f = strings.TrimSpace(f)
				if len(f) > 0 {
					res = append(res, f)
				}
			}

			return res
		}()
	} else {
		if len(*inFile) == 0 {
			log.Fatal("Input file or -all must be specified!")
		}

		files = append(files, *inFile)
	}

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