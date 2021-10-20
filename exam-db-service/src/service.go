package main

import (
	"context"
	"errors"
	pb "github.com/open-exam/open-exam-backend/exam-db-service/grpc-exam-db-service"
	"time"
)

type Server struct {
	pb.UnimplementedExamClientAccessServer
}

func NewServer() (*Server, error) {
	return &Server {}, nil
}

func (s *Server) CheckValid(ctx context.Context, req *pb.CheckValidRequest) (*pb.CheckValidResponse, error) {
	res := &pb.CheckValidResponse{}
	var id string

	if len(req.Id) != 0 {
		rows := db.QueryRow("SELECT * FROM exam_client_access WHERE id = ?", req.Id)
		if rows.Err() != nil {
			return nil, rows.Err()
		}

		if err := rows.Scan(&id, &res.UserId, &res.ExamId, &res.OpenAt, &res.ClosesAt); err != nil {
			return nil, err
		}

		now := time.Now().Unix()
		if res.OpenAt <= now && res.ClosesAt > now {
			res.Status = true
			return res, nil
		}

		res.Status = false
		return res, nil
	}
	return nil, errors.New("id not given")
}