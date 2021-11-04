package main

import (
	"context"
	"strconv"
	"strings"

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

func (s *Server) GetPlugins(ctx context.Context, req *pb.GetPluginsRequest) (*pb.PluginInfo, error) {
	items := []interface{}{req.Organization, req.BuildStatus}
	q := "SELECT * FROM plugins"
	args := []string{"organization = ?", "build_status = ?"}

	if len(req.Name) > 0 {
		args = append(args, "name LIKE ?")
		items = append(items, "%" + req.Name + "%")
	}

	if len(req.Uri) > 0 {
		args = append(args, "uri LIKE ?")
		items = append(items, "%" + req.Uri + "%")
	}

	if len(req.UriType) > 0 {
		args = append(args, "uri_type = ?")
		items = append(items, req.UriType)
	}

	if len(req.Version) > 0 {
		args = append(args, "version LIKE ?")
		items = append(items, "%" + req.Version + "%")
	}

	secondPart := strings.Join(args, " AND ")

	if len(secondPart) > 0 {
		secondPart = " WHERE " + secondPart
	}

	if req.NumPerPage <= 0 {
		req.NumPerPage = 25
	}

	rows, err := db.Query(q + secondPart + " LIMIT " + strconv.FormatInt(int64(req.NumPerPage), 10) + " OFFSET " + strconv.FormatInt(int64(req.NumPerPage * req.Page), 10), items...)
	if err != nil {
		return nil, err
	}

	var plugins []*pb.Plugin
	
	for rows.Next() {
		current := &pb.Plugin{}

		err := rows.Scan(&current.Id, &current.Name, &current.Uri, &current.UriType, &current.Version, &current.Organization, &current.BuildStatus)
		if err != nil {
			return nil, err
		}

		plugins = append(plugins, current)
	}

	countRows := db.QueryRow("SELECT COUNT(build_status) FROM plugins" + secondPart)
	if countRows.Err() != nil {
		return nil, countRows.Err()
	}

	var total int32
	err = countRows.Scan(&total)
	if err != nil {
		return nil, err
	}

	return &pb.PluginInfo {
		TotalItems: total,
		Plugins: plugins,
	}, nil
}