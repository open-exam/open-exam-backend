package main

import (
	"encoding/hex"
	"io/ioutil"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/open-exam/open-exam-backend/shared"
	"github.com/open-exam/open-exam-backend/util"
)

type Plugin struct {
	Name string `json:"name" binding:"required"`
	Uri string `json:"uri" binding:"required"`
	Version string `json:"version" binding:"required"`
	Organization uint64 `json:"organization" binding:"required"`
}

func InitPlugins(router *gin.RouterGroup) {
	router.Use(shared.JwtMiddleware(jwtPublicKey, mode))

	router.POST("/", addPlugin)
	router.GET("/", getPlugins)
}

func addPlugin(ctx *gin.Context) {
	var plugin Plugin
	if err := ctx.BindJSON(&plugin); err != nil {
		ctx.AbortWithStatusJSON(400, shared.GinErrors.JsonParseError)
		return
	}

	Id := hex.EncodeToString(util.GenerateRandomBytes(32))
	name, err := ioutil.TempDir("", Id)
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.UnknownError)
		return
	}

	log.Println(name)
	

	defer func() {
		os.RemoveAll(name)
	}()
}

func getPlugins(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "pong",
	})
}