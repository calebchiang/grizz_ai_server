package routes

import (
	"github.com/calebchiang/thirdparty_server/controllers"
	"github.com/calebchiang/thirdparty_server/middleware"
	"github.com/gin-gonic/gin"
)

func PracticeRoutes(r *gin.Engine) {

	auth := r.Group("/practice")
	auth.Use(middleware.RequireAuth())
	{
		auth.POST("/start", controllers.StartPractice)
		auth.GET("/ws", controllers.PracticeSocket)
		auth.POST("/finish", controllers.FinishPractice)
		auth.GET("/sessions", controllers.GetRecentActivity)
		auth.GET("/practice_overview", controllers.GetPracticeOverview)
		auth.GET("/skills_average", controllers.GetSkillsAverage)
	}
}
