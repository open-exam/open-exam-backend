package main

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	pb "github.com/open-exam/open-exam-backend/exam-db-service/grpc-exam-db-service"
	sharedPb "github.com/open-exam/open-exam-backend/grpc-shared"
	shared "github.com/open-exam/open-exam-backend/shared"
	util "github.com/open-exam/open-exam-backend/util"
)

type ExamClientAccessServer struct {
	pb.UnimplementedExamClientAccessServer
}

type ExamServiceServer struct {
	pb.UnimplementedExamServiceServer
}

type ExamTemplateServer struct {
	pb.UnimplementedExamTemplateServer
}

func NewExamClientAccessServer() (*ExamClientAccessServer, error) {
	return &ExamClientAccessServer {}, nil
}

func NewExamServiceServer() (*ExamServiceServer, error) {
	return &ExamServiceServer {}, nil
}

func NewExamTemplateServer() (*ExamTemplateServer, error) {
	return &ExamTemplateServer {}, nil
}

func (s *ExamClientAccessServer) CheckValid(ctx context.Context, req *pb.CheckValidRequest) (*pb.CheckValidResponse, error) {
	res := &pb.CheckValidResponse{}
	var id string

	if len(req.Id) != 0 {
		rows := db.QueryRow("SELECT * FROM exam_client_access WHERE id = ?", req.Id)
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

func (s *ExamServiceServer) CreateExam(ctx context.Context, req *pb.CreateExamRequest) (*sharedPb.StandardIdResponse, error) {
	{
		if len(req.Name) == 0 {
			return nil, errors.New("name not given")
		}
	
		if req.StartTime == 0 {
			return nil, errors.New("start time not given")
		}
	
		if req.EndTime == 0 {
			return nil, errors.New("end time not given")
		}
	
		if req.Duration == 0 {
			return nil, errors.New("duration not given")
		}
	
		if len(req.CreatedBy) == 0 {
			return nil, errors.New("created by not given")
		}
	
		if req.Organization == 0 {
			return nil, errors.New("organization not given")
		}

		if len(req.Scopes) == 0 {
			return nil, errors.New("scopes not given")
		}

		if len(req.Template) == 0 {
			return nil, errors.New("template not given")
		}
	}

	Id := hex.EncodeToString(util.GenerateRandomBytes(32))

	_, err := db.Exec("INSERT INTO exams VALUES (?, ?, ?, ?, ?, ?, ?, ?)", Id, req.Name, req.StartTime, req.EndTime, req.Duration, req.CreatedBy, req.Template, req.Organization)
	if err != nil {
		return nil, shared.Errors.UnknownError
	}

	for _, scope := range req.Scopes {
		_, err := db.Exec("INSERT INTO exam_scopes VALUES (?, ?, ?)", Id, scope.Scope, scope.ScopeType)
		if err != nil {
			return nil, err
		}
	}

	return &sharedPb.StandardIdResponse {
		IdString: Id,
	}, nil
}

func (s *ExamTemplateServer) CreateTemplate(ctx context.Context, req *pb.CreateExamTemplateRequest) (*sharedPb.StandardIdResponse, error) {
	{
		if len(req.Name) == 0 {
			return nil, errors.New("name not given")
		}
	
		if len(req.Template) == 0 {
			return nil, errors.New("template not given")
		}
	
		if len(req.Scopes) == 0 {
			return nil, errors.New("scopes not given")
		}
	}
	
	Id := hex.EncodeToString(util.GenerateRandomBytes(32))
	_, err := db.Exec("INSERT INTO exam_template VALUES (?, ?)", Id, req.Name)
	if err != nil {
		return nil, shared.Errors.UnknownError
	}

	for _, scope := range req.Scopes {
		_, err := db.Exec("INSERT INTO exam_template_scopes VALUES (?, ?, ?)", Id, scope.Scope, scope.ScopeType)
		if err != nil {
			return nil, err
		}
	}

	return &sharedPb.StandardIdResponse {
		IdString: Id,
	}, nil
}