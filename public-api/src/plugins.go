package main

import (
	"context"
	"encoding/hex"
	"io/ioutil"
	"os"

	"github.com/gin-gonic/gin"
	containerPb "github.com/open-exam/open-exam-backend/container-build-service/grpc-container-build-service"
	sharedPb "github.com/open-exam/open-exam-backend/grpc-shared"
	pluginDbPb "github.com/open-exam/open-exam-backend/plugin-db-service/grpc-plugin-db-service"
	"github.com/open-exam/open-exam-backend/shared"
	"github.com/open-exam/open-exam-backend/util"
)

type Plugin struct {
	Name         string `json:"name" binding:"required"`
	Uri          string `json:"uri" binding:"required"`
	UriType      string `json:"uri_type" binding:"required"`
	Version      string `json:"version" binding:"required"`
	Organization uint64 `json:"organization" binding:"required"`
}

type GetPlugins struct {
	Name        string `json:"name" binding:"required"`
	Uri         string `json:"uri" binding:"required"`
	UriType     string `json:"uri_type" binding:"required"`
	Version     string `json:"version" binding:"required"`
	BuildStatus bool   `json:"build_status" binding:"required"`
	Page        int32  `json:"page" binding:"required"`
	NumPerPage  int32  `json:"num_per_page" binding:"required"`
}

func InitPlugins(router *gin.RouterGroup) {
	router.Use(shared.JwtMiddleware(jwtPublicKey, mode))

	router.POST("/", shared.RBACMiddleware(rbacService, "PLUGINS", []string{"CREATE"}), addPlugin)
	router.GET("/", getPlugins)
}

func addPlugin(ctx *gin.Context) {
	var plugin Plugin
	if err := ctx.BindJSON(&plugin); err != nil {
		ctx.AbortWithStatusJSON(400, err)
		return
	}

	Id := hex.EncodeToString(util.GenerateRandomBytes(32))
	name, err := ioutil.TempDir("", Id)
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.UnknownError)
		return
	}

	resp, err := shared.GetPluginSources(Id, name, plugin.Uri, plugin.UriType)
	if err != nil {
		ctx.AbortWithStatusJSON(resp, gin.H{
			"error": err.Error(),
		})
		return
	}

	if _, err := shared.ParsePlugin(name); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": "could not validate plugin: " + err.Error(),
		})
		return
	}

	conn, err := shared.GetGrpcConn(pluginDbService)
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.ServiceConnection)
		return
	}

	client := pluginDbPb.NewPluginServiceClient(conn)

	res, err := client.AddPlugin(context.Background(), &pluginDbPb.Plugin{
		Name:         plugin.Name,
		Uri:          plugin.Uri,
		UriType:      plugin.UriType,
		Version:      plugin.Version,
		Organization: plugin.Organization,
	})
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.UnknownError)
		return
	}

	containerClient := containerPb.NewContainerClient(conn)
	resBuild, err := containerClient.Build(context.Background(), &sharedPb.StandardIdRequest{
		IdInt: res.IdInt,
	})
	if err != nil {
		ctx.AbortWithStatusJSON(500, err)
		return
	}

	if !resBuild.Status {
		ctx.AbortWithStatusJSON(400, gin.H{
			"error": resBuild.Message,
		})
		return
	}

	ctx.JSON(200, gin.H{
		"id": res.IdInt,
	})

	defer os.RemoveAll(name)
}

func getPlugins(ctx *gin.Context) {
	getPlugins := &GetPlugins{}
	if err := ctx.BindJSON(&getPlugins); err != nil {
		ctx.AbortWithStatusJSON(400, err)
		return
	}

	conn, err := shared.GetGrpcConn(pluginDbService)
	if err != nil {
		ctx.AbortWithStatusJSON(500, shared.GinErrors.ServiceConnection)
		return
	}

	client := pluginDbPb.NewPluginServiceClient(conn)
	res, err := client.GetPlugins(context.Background(), &pluginDbPb.GetPluginsRequest{
		Name:        getPlugins.Name,
		Uri:         getPlugins.Uri,
		UriType:     getPlugins.UriType,
		Version:     getPlugins.Version,
		BuildStatus: getPlugins.BuildStatus,
		Page:        getPlugins.Page,
		NumPerPage:  getPlugins.NumPerPage,
	})

	if err != nil {
		ctx.AbortWithStatusJSON(400, err)
		return
	}

	ctx.JSON(200, res)
}
