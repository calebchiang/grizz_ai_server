package routes

import (
	"github.com/calebchiang/thirdparty_server/controllers"
	"github.com/calebchiang/thirdparty_server/middleware"
	"github.com/gin-gonic/gin"
)

func SpeakingDrillRoutes(r *gin.Engine) {

	auth := r.Group("/drill")
	auth.Use(middleware.RequireAuth())
	{
		auth.POST("/start", controllers.StartSpeakingDrill)
	}
}
