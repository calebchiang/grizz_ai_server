package routes

import (
	"github.com/calebchiang/thirdparty_server/controllers"
	"github.com/calebchiang/thirdparty_server/middleware"
	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	r.POST("/users", controllers.CreateUser)
	r.POST("/login", controllers.LoginUser)
	r.POST("/apple_login", controllers.AppleLogin)

	auth := r.Group("/users")
	auth.Use(middleware.RequireAuth())
	{
		auth.GET("/me", controllers.GetCurrentUser)
		auth.PATCH("/name", controllers.UpdateUserName)
		auth.POST("/xp", controllers.AddXP)
		auth.GET("/weekly_overview", controllers.GetWeeklyOverview)
		auth.GET("/recent_practice", controllers.GetRecentPractice)
		auth.GET("/practice_challenge_overview", controllers.GetPracticeChallengeOverview)
		auth.POST("/seen_onboarding", controllers.MarkSeenOnboarding)
		auth.POST("/seen_ai_data_disclosure", controllers.MarkSeenAIDataDisclosure)
		auth.DELETE("/me", controllers.DeleteUser)
	}
}
