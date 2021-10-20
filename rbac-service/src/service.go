package main

import (
	"context"
	"database/sql"
	"errors"

	sharedPb "github.com/open-exam/open-exam-backend/grpc-shared"
	pb "github.com/open-exam/open-exam-backend/rbac-service/grpc-rbac-service"
)

var (
	operationDoesNotExist = &sharedPb.StandardStatusResponse {
		Status: false,
	}
)

type Server struct {
	pb.UnimplementedRbacServiceServer
}

func NewServer() (*Server, error) {
	return &Server{}, nil
}

func (s *Server) CanPerformOperation(ctx context.Context, req *pb.CanPerformOperationRequest) (*sharedPb.StandardStatusResponse, error) {

	var (
		rows      *sql.Row
		Id uint64
	)

	if req.OperationId > 0 {
		rows = db.QueryRow("SELECT id FROM operations WHERE id=?", req.OperationId)
		if rows.Err() != nil {
			return nil, rows.Err()
		}

		if err := rows.Scan(&Id); err != nil {
			return handleError(err)
		}
	} else {
		if len(req.Resource) == 0 || len(req.Operation) == 0 {
			return nil, errors.New("resource and operation is required")
		}

		rows = db.QueryRow("SELECT id FROM operations WHERE resource=? AND operation=?", req.Resource, req.Operation)
		if rows.Err() != nil {
			return nil, rows.Err()
		}

		if err := rows.Scan(&Id); err != nil {
			return handleError(err)
		}
	}

	rbacRows := db.QueryRow("SELECT id FROM rbac WHERE user_id=? AND oper_id=? AND scope=?", req.UserId, req.OperationId, req.Scope)

	if rbacRows.Err() != nil {
		return nil, rbacRows.Err()
	}

	err := rbacRows.Scan(&Id)
	
	if err != nil {
		return handleError(err)
	}

	return &sharedPb.StandardStatusResponse {
		Status: true,
	}, nil
}

func handleError(err error) (*sharedPb.StandardStatusResponse, error) {
	if err == sql.ErrNoRows {
		return operationDoesNotExist, nil
	}
	return nil, err
}