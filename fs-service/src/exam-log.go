package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	relationPb "github.com/open-exam/open-exam-backend/relation-service/grpc-relation-service"
	"github.com/open-exam/open-exam-backend/shared"
)

type LogData struct {
	UserId string `json:"user_id"`
	ExamId string `json:"exam_id"`
	Data string `json:"data"`
}

func InitExamLog(router *gin.RouterGroup) {
	router.PATCH("/", appendLog)
	router.GET("/", getUserLog)
}

func appendLog(ctx *gin.Context) {
	var log LogData

	if ctx.BindJSON(&log) == nil {
		if len(log.UserId) == 0 {
			dataNotGiven("user_id", ctx)
			return
		}

		if len(log.ExamId) == 0 {
			dataNotGiven("exam_id", ctx)
			return
		}

		if len(log.Data) == 0 {
			dataNotGiven("data", ctx)
			return
		}

		loggedAt := time.Now().Unix()

		conn, err := shared.GetGrpcConn("relation-service:" + relationService)

		if err != nil {
			ctx.JSON(500, shared.GinErrors.ServiceConnection)
			return
		}

		client := relationPb.NewRelationServiceClient(conn)
		res, err := GetOrgFromExam(client, log.ExamId)

		if err != nil {
			ctx.JSON(500, shared.GinErrors.ServiceConnection)
			return
		}

		rescanAccessExam, err := client.CanAccessExam(context.Background(), &relationPb.CanAccessExamRequest{
			UserId: log.UserId,
			ExamId: log.ExamId,
			VerifyTime: true,
		})

		if err != nil {
			ctx.JSON(500, shared.GinErrors.ServiceConnection)
			return
		}

		defer conn.Close()

		if !rescanAccessExam.Status {
			ctx.Status(403)
			return
		}

		logFilePath := "/app-data/exam-logs/" + strconv.FormatUint(res,
			10) + "/" + log.ExamId + "/" + log.UserId + ".log"

		f, err := os.OpenFile(logFilePath , os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			ctx.JSON(500, unknownError)
			return
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(len(log.Data)))
		binary.Write(buf, binary.BigEndian, loggedAt)
		binary.Write(buf, binary.BigEndian, log.Data)
		_, err = f.Write(buf.Bytes())
		if err != nil {
			ctx.JSON(500, unknownError)
			return
		}

		ctx.Status(200)

	} else {
		ctx.JSON(500, unknownError)
	}
}

func getUserLog(ctx *gin.Context) {
	var log LogData

	if ctx.BindJSON(&log) == nil {
		if len(log.UserId) == 0 {
			dataNotGiven("user_id", ctx)
			return
		}

		if len(log.ExamId) == 0 {
			dataNotGiven("exam_id", ctx)
			return
		}

		conn, err := shared.GetGrpcConn("exam-db-service:" + examDbService)

		if err != nil {
			ctx.JSON(500, shared.GinErrors.ServiceConnection)
			return
		}

		client := relationPb.NewRelationServiceClient(conn)
		res, err := GetOrgFromExam(client, log.ExamId)

		if err != nil {
			ctx.JSON(500, shared.GinErrors.ServiceConnection)
			return
		}

		ctx.File("/app-data/exam-logs/" + strconv.FormatUint(res, 10) + "/" + log.ExamId + "/" + log.UserId + ".log")
	} else {
		ctx.JSON(500, unknownError)
	}
}

func dataNotGiven(text string, ctx *gin.Context) {
	ctx.JSON(400, gin.H {
		"error": text + " not given",
	})
}