package main

import (
	"context"
	"github.com/gin-gonic/gin"
	examClientPb "github.com/open-exam/open-exam-backend/exam-client-access-service/grpc-exam-client-access-service"
	relationPb "github.com/open-exam/open-exam-backend/relation-service/grpc-relation-service"
	"github.com/open-exam/open-exam-backend/shared"
	"github.com/open-exam/open-exam-backend/util"
	"strconv"
)

var (
	unknownError = gin.H {
		"error": "an unknown error occurred",
	}
	platforms []string
	arch []string
	examClientAccessService string
	relationService string
	clientName string
)

type GetClient struct {
	Platform string `form:"platform"`
	Arch string `form:"arch"`
	AccessId string `form:"access_id"`
}

func InitExamFiles(router *gin.RouterGroup) {
	router.GET("/exam-client", getExamClient)
}

func getExamClient(ctx *gin.Context) {
	var getClient GetClient

	if ctx.BindQuery(&getClient) == nil {
		if len(getClient.Platform) == 0 || util.IsInList(getClient.Platform, &platforms) == -1 {
			ctx.JSON(400, gin.H {
				"error": "invalid platform",
			})
			return
		}

		if len(getClient.Arch) == 0 || util.IsInList(getClient.Arch, &arch) == -1 {
			ctx.JSON(400, gin.H {
				"error": "invalid arch",
			})
			return
		}

		if len(getClient.AccessId) == 0 {
			ctx.JSON(400, gin.H {
				"error": "access id is required",
			})
			return
		}

		conn, err := shared.GetGrpcConn("exam-client-access-service:" + examClientAccessService)

		if err != nil {
			ctx.JSON(500, errServiceConnection)
			return
		}

		client := examClientPb.NewExamClientAccessClient(conn)
		res, err := client.CheckValid(context.Background(), &examClientPb.CheckValidRequest{
			Id: getClient.AccessId,
		})

		defer conn.Close()

		if err != nil {
			ctx.JSON(500, errServiceConnection)
			return
		}

		// TODO: add logging

		if !res.Status {
			ctx.Status(403)
			return
		}

		orgId, err := getOrgFromExam(res.ExamId)

		if err != nil {
			ctx.JSON(500, unknownError)
			return
		}

		// TODO: perform any bit manipulation

		ctx.File("/app-data/" + strconv.FormatUint(orgId, 10) + "/client-files/" + res.ExamId + "/" + clientName + "_" + getClient.Arch + getExtension(getClient.Platform))
	} else {
		ctx.JSON(500, unknownError)
	}
}

func getOrgFromExam(examId string) (uint64, error) {
	conn, err := shared.GetGrpcConn("relation-service:" + relationService)
	if err != nil {
		return 0, err
	}

	client := relationPb.NewRelationServiceClient(conn)
	res, err := client.FindExamOrganisation(context.Background(), &relationPb.StandardIdRequest{
		IdString: examId,
	})

	if err != nil {
		return 0, err
	}

	return res.IdInt, nil
}

func getExtension(platform string) string {
	switch platform {
	case "win":
		return "exe"
	}
	return "exe"
}