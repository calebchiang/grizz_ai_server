package routes

import (
	"github.com/calebchiang/thirdparty_server/controllers"
	"github.com/calebchiang/thirdparty_server/middleware"
	"github.com/gin-gonic/gin"
)

func VideoRoutes(r *gin.Engine) {

	auth := r.Group("/video")
	auth.Use(middleware.RequireAuth())
	{
		auth.POST("/upload", controllers.UploadVideo)
	}
}
