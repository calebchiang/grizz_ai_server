package routes

import (
	"github.com/calebchiang/thirdparty_server/controllers"
	"github.com/calebchiang/thirdparty_server/middleware"
	"github.com/gin-gonic/gin"
)

func VocabularyRoutes(r *gin.Engine) {

	auth := r.Group("/vocabulary")
	auth.Use(middleware.RequireAuth())
	{
		auth.GET("", controllers.GetDailyVocabulary)
		auth.POST("/complete", controllers.CompleteVocabularySession)
		auth.POST("/progress", controllers.UpdateVocabularyProgress)
	}
}
