package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/calebchiang/thirdparty_server/services"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
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

	// ---------------------------
	// LOAD USER
	// ---------------------------

	var user models.User

	if err := database.DB.First(&user, userID.(uint)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load user",
		})
		return
	}

	// ---------------------------
	// CREDIT CHECK
	// ---------------------------

	if user.Credits <= 0 {

		// Premium user hit monthly limit
		if user.IsPremium {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"code":  "PREMIUM_LIMIT_REACHED",
				"error": "Monthly usage limit reached",
			})
		} else {
			// Free user ran out of credits
			c.JSON(http.StatusPaymentRequired, gin.H{
				"code":  "NO_CREDITS_FREE",
				"error": "No credits remaining",
			})
		}

		return
	}

	// ---------------------------
	// DEDUCT CREDIT
	// ---------------------------

	user.Credits -= 1

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update credits",
		})
		return
	}

	// ---------------------------
	// CREATE SESSION
	// ---------------------------

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

	// ---------------------------
	// RESPONSE
	// ---------------------------

	c.JSON(http.StatusCreated, gin.H{
		"session_id": session.ID,
		"scenario":   session.Scenario,
		"persona":    session.Persona,
		"started_at": session.StartedAt,
		"credits":    user.Credits,
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

	var strengths []string
	var weaknesses []string

	if err != nil {

		println("Failed to generate conversation result:", err.Error())

	} else {

		session.Clarity = result.Scores.Clarity
		session.Engagement = result.Scores.Engagement
		session.Confidence = result.Scores.Confidence
		session.ConversationFlow = result.Scores.ConversationFlow
		session.SocialAwareness = result.Scores.SocialAwareness

		session.ConversationScore = result.ConversationScore

		strengths = result.Strengths
		weaknesses = result.Weaknesses

		// Convert strengths to JSON
		strengthsJSON, err := json.Marshal(result.Strengths)
		if err == nil {
			session.Strengths = datatypes.JSON(strengthsJSON)
		}

		// Convert weaknesses to JSON
		weaknessesJSON, err := json.Marshal(result.Weaknesses)
		if err == nil {
			session.Weaknesses = datatypes.JSON(weaknessesJSON)
		}
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

		"strengths":  strengths,
		"weaknesses": weaknesses,

		"transcript": session.Transcript,
	})
}

func GetRecentActivity(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// ---------------------------
	// READ QUERY PARAM
	// ---------------------------

	limit := 0 // default = return everything

	if query := c.Query("limit"); query != "" {
		var parsed int
		_, err := fmt.Sscan(query, &parsed)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}

	type Activity struct {
		Type      string                 `json:"type"`
		CreatedAt time.Time              `json:"created_at"`
		Data      map[string]interface{} `json:"data"`
	}

	var practiceSessions []models.PracticeSession
	var speakingDrills []models.SpeakingDrill

	// ---------------------------
	// FETCH PRACTICE SESSIONS
	// ---------------------------

	queryPractice := database.DB.
		Where("user_id = ?", userID).
		Order("created_at desc")

	if limit > 0 {
		queryPractice = queryPractice.Limit(limit)
	}

	err := queryPractice.Find(&practiceSessions).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch practice sessions",
		})
		return
	}

	// ---------------------------
	// FETCH SPEAKING DRILLS
	// ---------------------------

	queryDrills := database.DB.
		Where("user_id = ?", userID).
		Order("created_at desc")

	if limit > 0 {
		queryDrills = queryDrills.Limit(limit)
	}

	err = queryDrills.Find(&speakingDrills).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch speaking drills",
		})
		return
	}

	var activities []Activity

	// ---------------------------
	// ADD PRACTICE SESSIONS
	// ---------------------------

	for _, session := range practiceSessions {

		activities = append(activities, Activity{
			Type:      "practice_session",
			CreatedAt: session.CreatedAt,
			Data: map[string]interface{}{
				"id":                 session.ID,
				"scenario":           session.Scenario,
				"persona":            session.Persona,
				"duration_seconds":   session.DurationSeconds,
				"started_at":         session.StartedAt,
				"ended_at":           session.EndedAt,
				"transcript":         session.Transcript,
				"clarity":            session.Clarity,
				"engagement":         session.Engagement,
				"confidence":         session.Confidence,
				"conversation_flow":  session.ConversationFlow,
				"social_awareness":   session.SocialAwareness,
				"conversation_score": session.ConversationScore,
				"strengths":          session.Strengths,
				"weaknesses":         session.Weaknesses,
				"created_at":         session.CreatedAt,
			},
		})
	}

	// ---------------------------
	// ADD SPEAKING DRILLS
	// ---------------------------

	for _, drill := range speakingDrills {

		activities = append(activities, Activity{
			Type:      "speaking_drill",
			CreatedAt: drill.CreatedAt,
			Data: map[string]interface{}{
				"id":             drill.ID,
				"topic":          drill.Topic,
				"transcript":     drill.Transcript,
				"video_url":      drill.VideoURL,
				"clarity":        drill.Clarity,
				"articulation":   drill.Articulation,
				"filler_rate":    drill.FillerRate,
				"pace":           drill.Pace,
				"structure":      drill.Structure,
				"speaking_score": drill.SpeakingScore,
				"filler_words":   drill.FillerWords,
				"strengths":      drill.Strengths,
				"weaknesses":     drill.Weaknesses,
				"phrases":        drill.PhrasesToUseInstead,
				"created_at":     drill.CreatedAt,
			},
		})
	}

	// ---------------------------
	// SORT ACTIVITIES
	// ---------------------------

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].CreatedAt.After(activities[j].CreatedAt)
	})

	// ---------------------------
	// FINAL LIMIT
	// ---------------------------

	if limit > 0 && len(activities) > limit {
		activities = activities[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
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

func GetSkillsAverage(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	type SkillsAverage struct {
		Clarity          float64 `json:"clarity"`
		Engagement       float64 `json:"engagement"`
		Confidence       float64 `json:"confidence"`
		ConversationFlow float64 `json:"conversation_flow"`
		SocialAwareness  float64 `json:"social_awareness"`
	}

	var result SkillsAverage

	err := database.DB.
		Model(&models.PracticeSession{}).
		Select(`
			AVG(clarity) as clarity,
			AVG(engagement) as engagement,
			AVG(confidence) as confidence,
			AVG(conversation_flow) as conversation_flow,
			AVG(social_awareness) as social_awareness
		`).
		Where("user_id = ? AND ended_at IS NOT NULL", userID).
		Scan(&result).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to calculate skill averages",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
