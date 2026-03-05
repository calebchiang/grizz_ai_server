package main

import (
	"os"

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
	)

	// ensure audio folder exists
	os.MkdirAll("audio", os.ModePerm)

	r := gin.Default()

	routes.UserRoutes(r)
	routes.PracticeRoutes(r)

	// serve generated TTS audio
	r.Static("/audio", "./audio")

	r.Run()
}
