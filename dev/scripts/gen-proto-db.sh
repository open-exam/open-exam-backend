#!/bin/bash
OUT=.
IN=user-service/grpc-user-service/user-service.proto
mkdir -p $OUT
protoc --go_out=$OUT --go_opt=paths=source_relative --go-grpc_out=$OUT --go-grpc_opt=paths=source_relative $IN