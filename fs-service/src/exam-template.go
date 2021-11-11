package main

import (
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/open-exam/open-exam-backend/shared"
)

func InitExamTemplate(router *gin.RouterGroup) {
	router.POST("/", addExamTemplate)
}

func addExamTemplate(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	files := form.File["template"]
	orgIds := form.Value["org_id"]

	if len(files) == 0 {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "no template file",
		})
		return
	}

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

	buf := make([]byte, files[0].Size)
	file, err := files[0].Open()
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": shared.GinErrors.UnknownError,
		})
		return
	}

	n, err := file.Read(buf)
	if err != nil || n == 0 {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": shared.GinErrors.UnknownError,
		})
		return
	}

	err = os.WriteFile("/app-data/exam-templates/"+orgIds[0]+"/"+files[0].Filename, buf, 0644)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": shared.GinErrors.UnknownError,
		})
		return
	}

	ctx.Status(200)
}
