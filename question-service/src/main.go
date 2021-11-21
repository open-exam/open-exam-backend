package src

import (
	"database/sql"
	"github.com/go-redis/redis/v8"
	pb "github.com/open-exam/open-exam-backend/question-service/grpc-question-service"
	"github.com/open-exam/open-exam-backend/shared"
	"github.com/open-exam/open-exam-backend/util"
	"google.golang.org/grpc"
	"os"
)

var (
	mode         string
	db           *sql.DB
	redisCluster *redis.ClusterClient
)

func main() {
	shared.SetEnv(&mode)

	validateOptions()

	shared.DefaultGrpcServer(func(server *grpc.Server) {
		db = shared.Db
		s, _ := NewServer()

		pb.RegisterQuestionServiceServer(server, s)
		pb.RegisterPoolServiceServer(server, s)
	})
}

func validateOptions() {
	redisAddrs := util.SplitAndParse(os.Getenv("redis_addrs"))
	redisPassword := os.Getenv("redis_pass")

	redisCluster = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    redisAddrs,
		Password: redisPassword,
	})
}
