package main

import (
	"context"

	sharedPb "github.com/open-exam/open-exam-backend/grpc-shared"
	pb "github.com/open-exam/open-exam-backend/plugin-db-service/grpc-plugin-db-service"
)

type Server struct {
	pb.UnimplementedPluginServiceServer
}

func NewServer() (*Server, error) {
	return &Server {}, nil
}

func (s* Server) AddPlugin(ctx context.Context, req *pb.Plugin) (*sharedPb.StandardIdResponse, error) {
	_, err := db.Exec("INSERT INTO plugins VALUES(?, ?, ?, ?, ?, ?)", req.Name, req.Uri, req.UriType, req.Version, req.Organization, false)

	if err != nil {
		return nil, err
	}

	var Id uint64
	res := db.QueryRow("SELECT id FROM plugins WHERE name=? AND uri=? AND uri_type=? AND version=? AND organization=?", req.Name, req.Uri, req.UriType, req.Version, req.Organization)
	if res.Err() != nil {
		return nil, err
	}
	
	err = res.Scan(&Id)
	if err != nil {
		return nil, err
	}

	return &sharedPb.StandardIdResponse{
		IdInt: Id,
	}, nil
}

func (s* Server) UpdateStatus(ctx context.Context, req *pb.UpdateStatus) (*sharedPb.StandardStatusResponse, error) {
	_, err := db.Exec("UPDATE plugins SET build_status = ? WHERE id = ?", req.Status, req.Id)
	if err != nil {
		return &sharedPb.StandardStatusResponse{
			Status: false,
		}, nil
	}

	return &sharedPb.StandardStatusResponse{
		Status: true,
	}, nil
}