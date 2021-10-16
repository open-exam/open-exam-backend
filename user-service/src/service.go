package main

import (
	"context"
	_ "database/sql"
	"errors"
	pb "github.com/open-exam/open-exam-backend/user-service/grpc-user-service"
)

type Server struct {
	pb.UnimplementedUserServiceServer
}

func NewServer() (*Server, error) {
	return &Server{}, nil
}

func (s *Server) FindOne(ctx context.Context, req *pb.FindOneRequest) (*pb.User, error) {
	if len(req.Id) != 0 {
		user, err := getUser(0, req.Id, req.Password)
		if err != nil {
			return nil, err
		}
		return user, nil
	} else if len(req.Email) != 0 {
		user, err := getUser(1, req.Email, req.Password)
		if err != nil {
			return nil, err
		}
		return user, nil
	}
	return nil, errors.New("id or email not given")
}

func getUser(mode int, data string, getPass bool) (*pb.User, error) {
	user := &pb.User{}

	query := "SELECT id, email, type"

	if getPass {
		query += ", password "
	}
	query += "FROM users WHERE "

	if mode == 0 {
		query += "id"
	} else {
		query += "email"
	}
	query += "=?"

	rows, err := db.Query(query, data)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if getPass {
			if err := rows.Scan(&user.Id, &user.Email, &user.Type, &user.Password); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(&user.Id, &user.Email, &user.Type); err != nil {
				return nil, err
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return user, nil
}