package main

import (
	"context"
	"errors"
	pb "github.com/open-exam/open-exam-backend/relation-service/grpc-relation-service"
)

type Server struct {
	pb.UnimplementedRelationServiceServer
}

func NewServer() (*Server, error) {
	return &Server {}, nil
}

func (s *Server) FindExamOrganisation(ctx context.Context, req *pb.StandardIdRequest) (*pb.StandardIdResponse, error) {
	var (
		scope uint64
		scopeType uint
	)

	if len(req.IdString) != 0 {
		rows := db.QueryRow("SELECT scope, scope_type FROM standard_users INNER JOIN exams ON standard_users.user_id = exams.created_by WHERE exams.id=? ORDER BY standard_users.scope_type", req.IdString)
		if rows.Err() != nil {
			return nil, rows.Err()
		}

		if err := rows.Scan(&scope, &scopeType); err == nil {
			if scopeType == 0 {
				return &pb.StandardIdResponse{
					IdInt: scope,
				}, nil
			} else {

				fillScope := func() (*pb.StandardIdResponse, error) {
					if rows.Err() != nil {
						return nil, rows.Err()
					}
					if err := rows.Scan(&scope); err == nil {
						return &pb.StandardIdResponse{
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