package src

import (
	"context"
	"errors"
	sharedPb "github.com/open-exam/open-exam-backend/grpc-shared"
	pb "github.com/open-exam/open-exam-backend/question-service/grpc-question-service"
)

type Server struct {
	pb.UnimplementedQuestionServiceServer
}

func NewServer() (*Server, error) {
	return &Server{}, nil
}

func (s *Server) getQuestionById(ctx context.Context, req *pb.QuestionById) (*pb.Question, error) {
	res := &pb.Question{}
	if (req.QuestionId) != 0 {
		row := db.QueryRow("SELECT plugin_id, title, display_data, files FROM questions where id = ?", req.QuestionId)
		if err := row.Scan(&res.PluginId, &res.Title, &res.DisplayData, &res.Files); err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, errors.New("id not given")
}

func (s *Server) getQuestionFromPool(ctx context.Context, req *pb.QuestionByPoolId) (*pb.Question, error) {
	res := &pb.Question{}
	if (req.PoolId) != 0 {
		if (req.QuestionId) != 0 {
			row := db.QueryRow("SELECT plugin_id, title, display_data, files FROM questions where id = (SELECT question_id FROM pool_questions WHERE pool_id = ? AND question_id = ?)", req.PoolId, req.QuestionId)
			if err := row.Scan(&res.PluginId, &res.Title, &res.DisplayData, &res.Files); err != nil {
				return nil, err
			}
			return res, nil
		}
		return nil, errors.New("question id not given")
	}
	return nil, errors.New("pool id not given")
}

func (s *Server) updateQuestion(ctx context.Context, req *pb.QuestionDetails) (*sharedPb.StandardStatusResponse, error) {
	items := make([]interface{}, 0)
	query := "UPDATE questions SET "
	if (req.QuestionId) == 0 {
		return nil, errors.New("QuestionId not given")
	}
	if (len(req.Data.DisplayData)) != 0 {
		query += "display_data = ?,"
		items = append(items, req.Data.DisplayData)
	}
	if (len(req.Data.Title)) != 0 {
		query += "title = ?,"
		items = append(items, req.Data.Title)
	}
	if (req.Data.PluginId) != 0 {
		query += "plugin_id = ?,"
		items = append(items, req.Data.PluginId)
	}
	if (len(req.Data.Files)) != 0 {
		query += "files = ?,"
		items = append(items, req.Data.Files)
	}
	query = query[:len(query)-1]
	_, err := db.Exec(query, items...)
	if err != nil {
		return &sharedPb.StandardStatusResponse{Status: true}, nil
	}
	return nil, err
}
