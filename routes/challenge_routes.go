package routes

import (
	"github.com/calebchiang/thirdparty_server/controllers"
	"github.com/calebchiang/thirdparty_server/middleware"
	"github.com/gin-gonic/gin"
)

func ChallengeRoutes(r *gin.Engine) {

	auth := r.Group("/challenges")
	auth.Use(middleware.RequireAuth())
	{
		auth.GET("/today", controllers.GetTodayChallenges)
		auth.POST("/complete", controllers.CompleteChallenge)
	}
}
