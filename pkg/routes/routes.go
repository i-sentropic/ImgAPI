package routes

import "github.com/gin-gonic/gin"

func RegisterRoutes(server *gin.Engine) {
	server.POST("/upload", uploadImage())
	server.GET("/download/:imageId", serveImage())
	server.POST("/fetch", fetchImage())
	server.POST("/delete", deleteImage())
	server.POST("/transform", transformImage())
}
