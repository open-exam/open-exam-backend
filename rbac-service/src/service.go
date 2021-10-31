package main

import (
	"context"
	"database/sql"
	"errors"

	sharedPb "github.com/open-exam/open-exam-backend/grpc-shared"
	pb "github.com/open-exam/open-exam-backend/rbac-service/grpc-rbac-service"
	relationPb "github.com/open-exam/open-exam-backend/relation-service/grpc-relation-service"
	"github.com/open-exam/open-exam-backend/shared"
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

func (s *Server) DoesRoleExist(ctx context.Context, req *pb.RoleExistRequest) (*pb.RoleExistResponse, error) {
	if len(req.Operation) == 0 || len(req.Resource) == 0 {
		return nil, errors.New("required fields missing")
	}

	rows := db.QueryRow("SELECT id FROM operations WHERE operation=? AND resource=?", req.Operation, req.Resource)

	var OperationId uint64
	if err := rows.Scan(&OperationId); err != nil {
		if err == sql.ErrNoRows {
			return &pb.RoleExistResponse {
				Status: false,
			}, nil
		}

		return nil, err
	}

	return &pb.RoleExistResponse {
		Status: true,
		OperationId: OperationId,
	}, nil
}

func (s *Server) CanPerformOperation(ctx context.Context, req *pb.CanPerformOperationRequest) (*sharedPb.StandardStatusResponse, error) {

	Ids := make([]uint64, 0)

	if req.OperationId > 0 {
		rows := db.QueryRow("SELECT id FROM operations WHERE id=?", req.OperationId)

		Ids = append(Ids, 0)
		if err := rows.Scan(&Ids[0]); err != nil {
			return handleError(err)
		}
	} else {
		if len(req.Resource) == 0 || len(req.Operation) == 0 {
			return nil, errors.New("resource and operation is required")
		}

		rows, err := db.Query("SELECT id FROM operations WHERE resource=? AND operation IN (?)", req.Resource, req.Operation)
		if err != nil {
			return nil, err
		}

		count := 0
		for rows.Next() {
			Ids = append(Ids, 0)
			rows.Scan(&Ids[count])
			count++
		}
	}

	if len(Ids) == 0 {
		return handleError(sql.ErrNoRows)
	}

	rbacRows, err := db.Query("SELECT id FROM rbac WHERE user_id=? AND oper_id IN(?) AND scope=?", req.UserId, Ids, req.Scope)

	if err != nil {
		return nil, err
	}

	if !rbacRows.Next() {
		return handleError(sql.ErrNoRows)
	}

	return &sharedPb.StandardStatusResponse {
		Status: true,
	}, nil
}

func (s *Server) GiveRole(ctx context.Context, req *pb.GiveRoleRequest) (*sharedPb.StandardStatusResponse, error) {
	if len(req.UserId) == 0 || req.OperationId == 0 || req.Scope == 0 || req.ScopeType == 0 {
		return nil, errors.New("required fields missing")
	}

	if res, err := s.checkAccessValidity(ctx, req); res != nil || err != nil {
		return res, err
	}

	_, err := db.Exec("INSERT INTO rbac (user_id, oper_id, scope, scope_type) VALUES (?, ?, ?, ?)", req.UserId, req.OperationId, req.Scope, req.ScopeType)
	if err != nil {
		return nil, err
	}

	return &sharedPb.StandardStatusResponse {
		Status: true,
	}, nil
}

func (s *Server) RevokeRole(ctx context.Context, req *pb.GiveRoleRequest) (*sharedPb.StandardStatusResponse, error) {
	if len(req.UserId) == 0 || req.OperationId == 0 || req.Scope == 0 || req.ScopeType == 0 {
		return nil, errors.New("required fields missing")
	}

	if res, err := s.checkAccessValidity(ctx, req); res != nil || err != nil {
		return res, err
	}

	_, err := db.Exec("DELETE FROM rbac WHERE user_id=? AND oper_id=? AND scope=? AND scope_type=?", req.UserId,
		req.OperationId, req.Scope, req.ScopeType)
	if err != nil {
		return nil, err
	}

	return &sharedPb.StandardStatusResponse {
		Status: true,
	}, nil
}

func (s *Server) checkAccessValidity(ctx context.Context, req *pb.GiveRoleRequest) (*sharedPb.StandardStatusResponse,
	error) {
	conn, err := shared.GetGrpcConn(relationService)

	if err != nil {
		return nil, err
	}

	client := relationPb.NewRelationServiceClient(conn)
	res, err := client.CanAccessScope(context.Background(), &relationPb.CanAccessScopeRequest{
		UserId: req.UserId,
		Scope: req.Scope,
	})

	if err != nil {
		return nil, err
	}

	if !res.Status {
		return &sharedPb.StandardStatusResponse {
			Status: false,
		}, nil
	}

	canPerform, err := s.CanPerformOperation(ctx, &pb.CanPerformOperationRequest{
		UserId: req.UserId,
		Resource: "SCOPE_ROLES",
		Operation: []string{"CREATE"},
	})

	if err != nil {
		return nil, err
	}

	if !canPerform.Status {
		return &sharedPb.StandardStatusResponse	{
			Status: false,
			Message: "inadequate permissions",
		}, nil
	}

	return nil, nil
}

func handleError(err error) (*sharedPb.StandardStatusResponse, error) {
	if err == sql.ErrNoRows {
		return operationDoesNotExist, nil
	}
	return nil, err
}