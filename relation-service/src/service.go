package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sharedPb "github.com/open-exam/open-exam-backend/grpc-shared"
	pb "github.com/open-exam/open-exam-backend/relation-service/grpc-relation-service"
)

type Server struct {
	pb.UnimplementedRelationServiceServer
}

func NewServer() (*Server, error) {
	return &Server{}, nil
}

func (s *Server) FindExamOrganization(ctx context.Context, req *sharedPb.StandardIdRequest) (*sharedPb.
	StandardIdResponse, error) {
	var scope uint64

	if len(req.IdString) != 0 {
		rows := db.QueryRow("SELECT org FROM exams where id=?", req.IdString)
		if rows.Err() != nil {
			return nil, rows.Err()
		}

		if err := rows.Scan(&scope); err == nil {
			return &sharedPb.StandardIdResponse{
				IdInt: scope,
			}, nil
		} else {
			return nil, err
		}
	}

	return nil, errors.New("id not given")
}

func (s *Server) CanAccessExam(ctx context.Context, req *pb.CanAccessExamRequest) (*sharedPb.StandardStatusResponse, error) {
	if len(req.ExamId) == 0 || len(req.UserId) == 0 {
		return nil, errors.New("invalid request")
	}

	var (
		startTime int64
		endTime   int64
		scope     uint64
		scopeType uint
		teamId    uint64
	)

	rows := db.QueryRow("SELECT ex.start_time, ex.end_time, es.scope, es.scope_type FROM exams ex RIGHT JOIN exam_scopes es ON ex.id = es.exam_id WHERE ex.id=? ORDER BY es.scope_type DESC", req.ExamId)
	if err := rows.Scan(&startTime, &endTime, &scope, &scopeType); err != nil {
		return nil, err
	}

	rows = db.QueryRow("SELECT team_id FROM students WHERE id=?", req.UserId)
	if err := rows.Scan(&teamId); err != nil {
		return nil, err
	}

	fillScope := func() (*sharedPb.StandardStatusResponse, error) {
		falseRes := &sharedPb.StandardStatusResponse {
			Status: false,
		}
		res := &sharedPb.StandardStatusResponse {
			Status: scope == teamId,
		}

		if rows.Err() != nil {
			return nil, rows.Err()
		}
		if err := rows.Scan(&scope); err != nil {
			return falseRes, nil
		}

		now := time.Now().Unix()
		if req.VerifyTime {
			if now >= startTime && now <= endTime {
				return res, nil
			}

			return falseRes, nil
		}
		
		return res, nil
	}

	switch scopeType {
	case 3: {
		rows = db.QueryRow("SELECT tm.id FROM teams tm WHERE tm.id=? AND tm.group_id IN(SELECT gr.id FROM `groups` gr WHERE gr.org_id=?)", teamId, scope)
	}
	case 4: {
		rows = db.QueryRow("SELECT id FROM teams WHERE group_id=?", scope)
	}
	}
	return fillScope()
}


func (s *Server) CanAccessScope(ctx context.Context, req *pb.CanAccessScopeRequest) (*sharedPb.StandardStatusResponse, error) {
	
	if len(req.UserId) == 0 || req.Scope == 0 {
		return nil, errors.New("invalid request")
	}

	var (
		scope     uint64
		scopeType uint
		trueStatus = &sharedPb.StandardStatusResponse {
			Status: true,
		};
		innerRows *sql.Row
	)

	singleRow := db.QueryRow("SELECT scope, scope_type FROM standard_users WHERE user_id=? AND scope=?", req.UserId, req.Scope)

	if singleRow.Err() != nil && singleRow.Err() != sql.ErrNoRows {
		return nil, singleRow.Err()
	} else if singleRow.Err() == nil {
		return trueStatus, nil
	}

	rows, err := db.Query("SELECT scope, scope_type FROM standard_users WHERE user_id=? ORDER BY scope_type ASC", req.UserId)
	
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err := rows.Scan(&scope, &scopeType); err != nil {
			return nil, err
		}

		if scopeType == 1 || scope == req.Scope {
			return trueStatus, nil
		}

		switch scopeType {
		case 2: {
			if scope != req.Scope {
				continue
			}
		}
		case 3: {
			innerRows = db.QueryRow("SELECT id FROM `groups` WHERE org_id=? AND id=?", scope, req.Scope)
		}
		case 4: {
			innerRows = db.QueryRow("SELECT id FROM teams WHERE group_id=? AND id=?", scope, req.Scope)
		}
		}

		if innerRows.Err() != nil && innerRows.Err() != sql.ErrNoRows {
			return nil, innerRows.Err()
		} else if singleRow.Err() == nil {
			return trueStatus, nil
		}
	}

	return &sharedPb.StandardStatusResponse {
		Status: false,
	}, nil
}