package controllers

import (
	"net/http"
	"time"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/gin-gonic/gin"
)

func StartPractice(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var input struct {
		Scenario string `json:"scenario"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if input.Scenario == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Scenario is required",
		})
		return
	}

	session := models.PracticeSession{
		UserID:    userID.(uint),
		Scenario:  input.Scenario,
		StartedAt: time.Now(),
	}

	if err := database.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create practice session",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_id": session.ID,
		"scenario":   session.Scenario,
		"started_at": session.StartedAt,
	})
}
