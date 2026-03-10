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

	// Reconstruct transcript
	transcript := services.ReconstructTranscript(session.Persona, messages)

	now := time.Now()

	duration := int(now.Sub(session.StartedAt).Seconds())

	session.EndedAt = &now
	session.DurationSeconds = duration
	session.Transcript = transcript

	// Generate conversation result using OpenAI
	result, err := services.GenerateConversationResult(transcript)

	if err != nil {

		println("Failed to generate conversation result:", err.Error())

	} else {

		session.Clarity = result.Scores.Clarity
		session.Engagement = result.Scores.Engagement
		session.Confidence = result.Scores.Confidence
		session.ConversationFlow = result.Scores.ConversationFlow
		session.SocialAwareness = result.Scores.SocialAwareness

		session.ConversationScore = result.ConversationScore
	}

	if err := database.DB.Save(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to finish practice session",
		})
		return
	}

	// ---------------------------
	// ADD XP FOR PRACTICE
	// ---------------------------

	var user models.User

	if err := database.DB.First(&user, userID.(uint)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load user",
		})
		return
	}

	user.XP += 50

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update XP",
		})
		return
	}

	// Update streak
	services.UpdateUserStreak(userID.(uint))

	c.JSON(http.StatusOK, gin.H{
		"message":            "Practice session finished",
		"duration_seconds":   duration,
		"conversation_score": session.ConversationScore,

		"clarity":           session.Clarity,
		"engagement":        session.Engagement,
		"confidence":        session.Confidence,
		"conversation_flow": session.ConversationFlow,
		"social_awareness":  session.SocialAwareness,

		"transcript": session.Transcript,
	})
}

func GetPracticeSessions(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var sessions []models.PracticeSession

	err := database.DB.
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&sessions).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch practice sessions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
	})
}

func GetPracticeOverview(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	type SessionScore struct {
		Score int `json:"score"`
	}

	var sessions []models.PracticeSession

	err := database.DB.
		Select("conversation_score, created_at").
		Where("user_id = ? AND ended_at IS NOT NULL", userID).
		Order("created_at desc").
		Limit(30).
		Find(&sessions).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch sessions",
		})
		return
	}

	// Reverse sessions to oldest → newest
	for i, j := 0, len(sessions)-1; i < j; i, j = i+1, j-1 {
		sessions[i], sessions[j] = sessions[j], sessions[i]
	}

	var scores []SessionScore
	total := 0

	for _, s := range sessions {
		scores = append(scores, SessionScore{
			Score: s.ConversationScore,
		})
		total += s.ConversationScore
	}

	average := 0
	if len(scores) > 0 {
		average = total / len(scores)
	}

	// FIX: prevent null JSON array
	if scores == nil {
		scores = []SessionScore{}
	}

	c.JSON(http.StatusOK, gin.H{
		"average_score": average,
		"sessions":      scores,
	})
}
