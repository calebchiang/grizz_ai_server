package controllers

import (
	"net/http"
	"os"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/calebchiang/thirdparty_server/services"
	"github.com/gin-gonic/gin"
)

func StartSpeakingDrill(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	topic := c.PostForm("topic")

	if topic == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Topic is required",
		})
		return
	}

	file, err := c.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Audio file is required",
		})
		return
	}

	// create temp file path
	tempPath := "./tmp/" + file.Filename

	// save uploaded file
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save audio file",
		})
		return
	}

	// transcribe audio
	transcript, err := services.TranscribeAudio(tempPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to transcribe audio",
		})
		return
	}

	// remove temp file after transcription
	os.Remove(tempPath)

	drill := models.SpeakingDrill{
		UserID:     userID.(uint),
		Topic:      topic,
		Transcript: transcript,
	}

	if err := database.DB.Create(&drill).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create speaking drill",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Speaking drill created",
		"drill_id":   drill.ID,
		"topic":      drill.Topic,
		"transcript": drill.Transcript,
		"created_at": drill.CreatedAt,
	})
}
