package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"

	pb "github.com/open-exam/open-exam-backend/container-build-service/grpc-container-build-service"
	sharedPb "github.com/open-exam/open-exam-backend/grpc-shared"
	pluginDbPb "github.com/open-exam/open-exam-backend/plugin-db-service/grpc-plugin-db-service"
	"github.com/open-exam/open-exam-backend/shared"
	"github.com/open-exam/open-exam-backend/util"
)

type Server struct {
	pb.UnimplementedContainerServer
}

func NewServer() (*Server, error) {
	return &Server{}, nil
}

func (s *Server) Build(ctx context.Context, req *sharedPb.StandardIdRequest) (*sharedPb.StandardStatusResponse, error) {
	rows := db.QueryRow("SELECT uri, uri_type FROM plugins WHERE id=?", req.IdInt)
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var (
		uri     string
		uriType string
	)

	if err := rows.Scan(&uri, &uriType); err != nil {
		if err == sql.ErrNoRows {
			return &sharedPb.StandardStatusResponse{
				Status:  false,
				Message: "Plugin not found",
			}, nil
		}

		return nil, err
	}

	Id := hex.EncodeToString(util.GenerateRandomBytes(32))
	name, err := ioutil.TempDir("", Id)
	if err != nil {
		return nil, err
	}

	_, err = shared.GetPluginSources(Id, name, uri, uriType)
	if err != nil {
		return nil, err
	}

	plugin, err := shared.ParsePlugin(name)
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(name)

	if plugin.ServerSide != nil {
		defer func() {
			IdString := strconv.FormatUint(req.IdInt, 10)

			cmd := exec.Command("podman", "build", "-t", IdString, "-f", plugin.Dockerfile, plugin.Context)
			cmdErr := cmd.Wait()
			if cmdErr != nil {
				log.Println(err)
				goto setStatus
			}

			cmd = exec.Command("podman", "tag", "localhost/"+IdString+":latest", registryHost+":"+registryPort+"/"+"plugins/"+IdString+":latest")
			cmdErr = cmd.Wait()
			if cmdErr != nil {
				log.Println(err)
				goto setStatus
			}

			cmd = exec.Command("podman", "push", registryHost+":"+registryPort+"/"+"plugins/"+IdString+":latest")
			cmdErr = cmd.Wait()
			if cmdErr != nil {
				log.Println(err)
			}

		setStatus:
			conn, err := shared.GetGrpcConn(pluginDbService)
			if err != nil {
				log.Println(err)
				return
			}

			client := pluginDbPb.NewPluginServiceClient(conn)

			_, err = client.UpdateStatus(context.Background(), &pluginDbPb.UpdateStatus{
				Id:     req.IdInt,
				Status: cmdErr == nil,
			})

			if err != nil {
				log.Println(err)
			}
		}()
	}

	return &sharedPb.StandardStatusResponse{
		Status: true,
	}, nil
}
