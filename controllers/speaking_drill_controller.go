package controllers

import (
	"net/http"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
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

	var input struct {
		Topic      string `json:"topic"`
		Transcript string `json:"transcript"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if input.Topic == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Topic is required",
		})
		return
	}

	if input.Transcript == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Transcript is required",
		})
		return
	}

	drill := models.SpeakingDrill{
		UserID:     userID.(uint),
		Topic:      input.Topic,
		Transcript: input.Transcript,
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
