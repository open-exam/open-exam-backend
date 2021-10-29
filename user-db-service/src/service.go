package main

import (
	"context"
	"database/sql"
	_ "database/sql"
	"encoding/hex"
	"errors"
	"github.com/go-sql-driver/mysql"
	sharedPb "github.com/open-exam/open-exam-backend/grpc-shared"
	pb "github.com/open-exam/open-exam-backend/user-db-service/grpc-user-db-service"
	"github.com/open-exam/open-exam-backend/util"
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

func (s *Server) CreateUser(ctx context.Context, req *pb.User) (*pb.User, error) {
	if len(req.Email) == 0 {
		return nil, errors.New("email is required")
	}

	password := generatePassword(standardPasswordSize)
	passHash, err := generateFromPassword(password)
	if err != nil {
		return nil, errors.New("unknown error while generating password")
	}

	Id := hex.EncodeToString(util.GenerateRandomBytes(32))

	_, err = db.Exec("INSERT INTO users VALUES (?, ?, ?, ?, ?)", Id, req.Email, req.Type, passHash, req.Name)
	if err != nil {
		if err.(*mysql.MySQLError).Number == 1062 {
			rows := db.QueryRow("SELECT id FROM users WHERE email=?", req.Email)

			err = rows.Scan(&Id)
			if err != nil {
				return nil, err
			}

			return &pb.User {
				Id: Id,
			}, nil
		}

		return nil, err
	}

	return &pb.User {
		Id: Id,
	}, nil
}

func (s *Server) AddUserToScope(ctx context.Context, req *pb.AddUser) (*sharedPb.StandardStatusResponse, error) {
	if len(req.Id) == 0 {
		return nil, errors.New("id is required")
	}

	if req.Scope == 0 || req.ScopeType == 0 {
		return nil, errors.New("scope and scopeType are required")
	}

	row := db.QueryRow("SELECT type FROM users WHERE id=?", req.Id)

	var userType uint32
	err := row.Scan(&userType)
	if err != nil {
		if err == sql.ErrNoRows {
			return &sharedPb.StandardStatusResponse{
				Status: false,
				Message: "user does not exist",
			}, nil
		}

		return nil, err
	}

	if userType == 1 && req.ScopeType != 4 {
		return &sharedPb.StandardStatusResponse{
			Status: false,
			Message: "student can only be assigned to teams",
		}, nil
	}

	if userType == 1 {
		_, err = db.Exec("INSERT INTO students(id, team_id) VALUES (?,?)", req.Id, req.Scope)
	} else {
		_, err = db.Exec("INSERT INTO standard_users VALUES (?,?,?)", req.Id, req.Scope, req.ScopeType)
	}

	if err != nil {
		return nil, err
	}

	return &sharedPb.StandardStatusResponse {
		Status: true,
	}, nil
}