package routes

import (
	"github.com/calebchiang/thirdparty_server/controllers"
	"github.com/calebchiang/thirdparty_server/middleware"
	"github.com/gin-gonic/gin"
)

func TopicRoutes(r *gin.Engine) {

	auth := r.Group("/topics")
	auth.Use(middleware.RequireAuth())
	{
		auth.GET("", controllers.GetTopics)
	}
}
