package main

import (
	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/calebchiang/thirdparty_server/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	database.Connect()

	database.DB.AutoMigrate(
		&models.User{},
		&models.PracticeSession{},
		&models.PracticeMessage{},
		&models.Challenge{},
		&models.ChallengeCompletion{},
		&models.Topic{},
	)

	r := gin.Default()

	routes.UserRoutes(r)
	routes.PracticeRoutes(r)
	routes.ChallengeRoutes(r)
	routes.RevenueCatRoutes(r)

	r.Run()
}
