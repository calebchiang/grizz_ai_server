package routes

import (
	"github.com/calebchiang/thirdparty_server/controllers"
	"github.com/calebchiang/thirdparty_server/middleware"
	"github.com/gin-gonic/gin"
)

func CourseRoutes(r *gin.Engine) {

	auth := r.Group("/courses")
	auth.Use(middleware.RequireAuth())
	{
		auth.GET("/", controllers.GetCourses)
	}
}
