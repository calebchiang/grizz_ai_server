package controllers

import (
	"net/http"
	"time"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/calebchiang/thirdparty_server/services"
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
		Persona  string `json:"persona"`
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

	if input.Persona == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Persona is required",
		})
		return
	}

	session := models.PracticeSession{
		UserID:    userID.(uint),
		Scenario:  input.Scenario,
		Persona:   input.Persona,
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
		"persona":    session.Persona,
		"started_at": session.StartedAt,
	})
}

func FinishPractice(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var input struct {
		SessionID uint `json:"session_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	var session models.PracticeSession

	err := database.DB.
		Where("id = ? AND user_id = ?", input.SessionID, userID).
		First(&session).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Practice session not found",
		})
		return
	}

	// Load session messages
	var messages []models.PracticeMessage

	err = database.DB.
		Where("session_id = ?", session.ID).
		Order("created_at asc").
		Find(&messages).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load session messages",
		})
		return
	}

	transcript := services.ReconstructTranscript(session.Persona, messages)

	now := time.Now()

	duration := int(now.Sub(session.StartedAt).Seconds())

	session.EndedAt = &now
	session.DurationSeconds = duration
	session.Transcript = transcript

	if err := database.DB.Save(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to finish practice session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Practice session finished",
		"duration_seconds": duration,
	})
}
