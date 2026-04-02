package routes

import (
	"github.com/calebchiang/thirdparty_server/controllers"
	"github.com/calebchiang/thirdparty_server/middleware"
	"github.com/gin-gonic/gin"
)

func DictionaryRoutes(r *gin.Engine) {

	auth := r.Group("/dictionary")
	auth.Use(middleware.RequireAuth())
	{
		auth.POST("/save", controllers.SaveWord)
		auth.GET("/", controllers.GetUserDictionary)
		auth.GET("/recent_words", controllers.GetRecentWords)
	}
}
