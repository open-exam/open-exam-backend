package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/open-exam/open-exam-backend/shared"
)

func InitExamTemplates(router *gin.RouterGroup) {
	router.Use(shared.JwtMiddleware(jwtPublicKey, mode))

	router.POST("/", createExamTemplate)
}

func createExamTemplate(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": err.Error(),
		})
		return
	}

	orgIds := form.Value["org_id"]
	if len(orgIds) == 0 || len(orgIds[0]) == 0 {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "no org_id",
		})
		return
	}

	_, err = strconv.ParseUint(orgIds[0], 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "invalid organization id",
		})
		return
	}
	
	files := form.File["template"]
	if len(files) == 0 {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": "No template file provided",
		})
		return
	}
	
	buf := make([]byte, files[0].Size)
	file, err := files[0].Open()
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": shared.GinErrors.UnknownError,
		})
		return
	}

	n, err := file.Read(buf)
	if err != nil || n == 0 {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": shared.GinErrors.UnknownError,
		})
		return
	}

	examConfig := &shared.Exam{}
	err = json.Unmarshal(buf, examConfig)
	if err != nil {
		ctx.AbortWithStatusJSON(400, shared.GinErrors.JsonParseError)
		return
	}

	if err := examConfig.Validate(); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H {
			"error": err.Error(),
		})
		return
	}

	ct, body, err := createForm(map[string]string {
		"template": string(buf),
		"org_id": form.Value["org_id"][0],
	})
	resp, err := http.Post(fsService + "/exam-template", ct, body)
	if err != nil || resp.StatusCode >= 400 {
		ctx.AbortWithStatusJSON(400, shared.GinErrors.UnknownError)
		return
	}

	ctx.Status(200)
}

func createForm(form map[string]string) (string, io.Reader, error) {
	body := new(bytes.Buffer)
	mp := multipart.NewWriter(body)
	defer mp.Close()
	for key, val := range form {
		mp.WriteField(key, val)
	}
	return mp.FormDataContentType(), body, nil
}