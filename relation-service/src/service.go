package main

import (
	"context"
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
	var (
		scope     uint64
		scopeType uint
	)

	if len(req.IdString) != 0 {
		rows := db.QueryRow("SELECT scope, scope_type FROM standard_users INNER JOIN exams ON standard_users.user_id = exams.created_by WHERE exams.id=? ORDER BY standard_users.scope_type", req.IdString)
		if rows.Err() != nil {
			return nil, rows.Err()
		}

		if err := rows.Scan(&scope, &scopeType); err == nil {
			if scopeType == 0 {
				return &sharedPb.StandardIdResponse{
					IdInt: scope,
				}, nil
			} else {

				fillScope := func() (*sharedPb.StandardIdResponse, error) {
					if rows.Err() != nil {
						return nil, rows.Err()
					}
					if err := rows.Scan(&scope); err == nil {
						return &sharedPb.StandardIdResponse{
							IdInt: scope,
						}, nil
					}
					return nil, err
				}

				switch scopeType {
				case 1: {
					rows = db.QueryRow("SELECT org_id FROM `groups` WHERE id=?", scope)
				}
				case 2: {
					rows = db.QueryRow("SELECT org_id FROM `groups` INNER JOIN teams ON teams.group_id = `groups`.id WHERE teams.id=?", scope)
				}
				case 3: {
					return nil, errors.New("maps to a custom_team")
				}
				}
				return fillScope()
			}
		} else {
			return nil, err
		}
	}

	return nil, errors.New("id not given")
}

func (s *Server) HasAccess(ctx context.Context, req *pb.HasAccessRequest) (*sharedPb.StandardValidResponse, error) {
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

	fillScope := func() (*sharedPb.StandardValidResponse, error) {
		falseRes := &sharedPb.StandardValidResponse {
			Status: false,
		}
		res := &sharedPb.StandardValidResponse {
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
	case 0: {
		rows = db.QueryRow("SELECT tm.id FROM teams tm WHERE tm.id=? AND tm.group_id IN(SELECT gr.id FROM `groups` gr WHERE gr.org_id=?)", teamId, scope)
	}
	case 1: {
		rows = db.QueryRow("SELECT id FROM teams WHERE group_id=?", scope)
	}
	}
	return fillScope()
}
