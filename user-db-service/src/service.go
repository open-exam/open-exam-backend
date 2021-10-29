package main

import (
	"context"
	"database/sql"
	_ "database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	sharedPb "github.com/open-exam/open-exam-backend/grpc-shared"
	pb "github.com/open-exam/open-exam-backend/user-db-service/grpc-user-db-service"
	"github.com/open-exam/open-exam-backend/util"
	"io"
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

func (s *Server) CreateUser(stream pb.UserService_CreateUserServer) error {
	handleStreamSend := func(user *pb.User) {
		err := stream.Send(user)
		if err != nil {
			fmt.Println("could not send response")
		}
	}

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if len(req.Email) == 0 {
			handleStreamSend(&pb.User {
				Error: "email is required",
			})
			continue
		}

		password := generatePassword(standardPasswordSize)
		passHash, err := generateFromPassword(password)
		if err != nil {
			handleStreamSend(&pb.User {
				Error: "unknown error while generating password",
			})
			continue
		}

		Id := hex.EncodeToString(util.GenerateRandomBytes(32))

		_, err = db.Exec("INSERT INTO users VALUES (?, ?, ?, ?, ?)", Id, req.Email, req.Type, passHash, req.Name)
		if err != nil {
			if err.(*mysql.MySQLError).Number == 1062 {
				rows := db.QueryRow("SELECT id FROM users WHERE email=?", req.Email)

				err = rows.Scan(&Id)
				if err != nil {
					handleStreamSend(&pb.User {
						Error: "an unknown error occurred",
					})
					continue
				}

				handleStreamSend(&pb.User {
					Id: Id,
				})
				continue
			}

			handleStreamSend(&pb.User {
				Error: "an unknown error occurred",
			})
			continue
		}

		handleStreamSend(&pb.User {
			Id: Id,
		})
	}
}

func (s *Server) AddUserToScope(stream pb.UserService_AddUserToScopeServer) error {
	handleStreamSend := func(user *sharedPb.StandardStatusResponse) {
		err := stream.Send(user)
		if err != nil {
			fmt.Println("could not send response")
		}
	}

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if len(req.Id) == 0 {
			handleStreamSend(&sharedPb.StandardStatusResponse {
				Status: false,
				Message: "id is required",
			})
			continue
		}

		if req.Scope == 0 || req.ScopeType == 0 {
			handleStreamSend(&sharedPb.StandardStatusResponse {
				Status: false,
				Message: "scope and scopeType are required",
			})
			continue
		}

		row := db.QueryRow("SELECT type FROM users WHERE id=?", req.Id)

		var userType uint32
		err = row.Scan(&userType)
		if err != nil {
			if err == sql.ErrNoRows {
				handleStreamSend(&sharedPb.StandardStatusResponse {
					Status: false,
					Message: "user does not exist",
				})
				continue
			}

			handleStreamSend(&sharedPb.StandardStatusResponse{
				Status: false,
				Message: err.Error(),
			})
			continue
		}

		if userType == 1 && req.ScopeType != 4 {
			handleStreamSend(&sharedPb.StandardStatusResponse {
				Status: false,
				Message: "student can only be assigned to teams",
			})
			continue
		}

		if userType == 1 {
			_, err = db.Exec("INSERT INTO students(id, team_id) VALUES (?,?)", req.Id, req.Scope)
		} else {
			_, err = db.Exec("INSERT INTO standard_users VALUES (?,?,?)", req.Id, req.Scope, req.ScopeType)
		}

		if err != nil {
			handleStreamSend(&sharedPb.StandardStatusResponse {
				Status: false,
				Message: "an unknown error occurred",
			})
			continue
		}

		handleStreamSend(&sharedPb.StandardStatusResponse {
			Status: true,
		})
	}
}