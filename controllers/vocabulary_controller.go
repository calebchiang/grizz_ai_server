package controllers

import (
	"net/http"
	"time"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/gin-gonic/gin"
)

func GetDailyVocabulary(c *gin.Context) {

	db := database.DB

	// ---------------------------
	// GET USER FROM CONTEXT
	// ---------------------------

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID := userIDRaw.(uint)

	var user models.User

	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// ---------------------------
	// DETERMINE USER LOCAL DAY
	// ---------------------------

	location, err := time.LoadLocation(user.Timezone)
	if err != nil {
		location = time.UTC
	}

	now := time.Now().In(location)

	today := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		location,
	)

	// ---------------------------
	// FIND OR CREATE SESSION
	// ---------------------------

	var session models.VocabularySession

	err = db.
		Where("user_id = ? AND date = ?", user.ID, today).
		First(&session).Error

	if err != nil {

		session = models.VocabularySession{
			UserID: user.ID,
			Date:   today,
		}

		if err := db.Create(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create vocabulary session",
			})
			return
		}

		// ---------------------------
		// FETCH 5 RANDOM WORDS
		// ---------------------------

		var words []models.Vocabulary

		if err := db.
			Order("RANDOM()").
			Limit(5).
			Find(&words).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch vocabulary",
			})
			return
		}

		// ---------------------------
		// STORE SESSION WORDS
		// ---------------------------

		for i, word := range words {

			sessionWord := models.VocabularySessionWord{
				SessionID:    session.ID,
				VocabularyID: word.ID,
				OrderIndex:   i,
				Completed:    false,
			}

			db.Create(&sessionWord)
		}
	}

	// ---------------------------
	// LOAD SESSION WORDS
	// ---------------------------

	var sessionWords []models.VocabularySessionWord

	if err := db.
		Preload("Vocabulary").
		Where("session_id = ?", session.ID).
		Order("order_index ASC").
		Find(&sessionWords).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load session vocabulary",
		})
		return
	}

	// ---------------------------
	// BUILD RESPONSE WORD LIST
	// ---------------------------

	words := []models.Vocabulary{}

	for _, sw := range sessionWords {
		words = append(words, sw.Vocabulary)
	}

	if words == nil {
		words = []models.Vocabulary{}
	}

	// ---------------------------
	// CALCULATE CURRENT INDEX
	// ---------------------------

	currentIndex := 0

	for i, sw := range sessionWords {
		if !sw.Completed {
			currentIndex = i
			break
		}
	}

	// ---------------------------
	// RESPONSE
	// ---------------------------

	c.JSON(http.StatusOK, gin.H{
		"words":         words,
		"completed":     session.Completed,
		"current_index": currentIndex,
	})
}

func UpdateVocabularyProgress(c *gin.Context) {

	db := database.DB

	// ---------------------------
	// GET USER FROM CONTEXT
	// ---------------------------

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID := userIDRaw.(uint)

	var body struct {
		OrderIndex int `json:"order_index"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// ---------------------------
	// FIND USER
	// ---------------------------

	var user models.User

	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	location, err := time.LoadLocation(user.Timezone)
	if err != nil {
		location = time.UTC
	}

	now := time.Now().In(location)

	today := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		location,
	)

	// ---------------------------
	// FIND TODAY SESSION
	// ---------------------------

	var session models.VocabularySession

	if err := db.
		Where("user_id = ? AND date = ?", user.ID, today).
		First(&session).Error; err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "Session not found"})
		return
	}

	// ---------------------------
	// FIND SESSION WORD
	// ---------------------------

	var sessionWord models.VocabularySessionWord

	if err := db.
		Where("session_id = ? AND order_index = ?", session.ID, body.OrderIndex).
		First(&sessionWord).Error; err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "Word not found"})
		return
	}

	sessionWord.Completed = true

	db.Save(&sessionWord)

	c.JSON(http.StatusOK, gin.H{
		"message": "Progress updated",
	})
}

func CompleteVocabularySession(c *gin.Context) {

	db := database.DB

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID := userIDRaw.(uint)

	var user models.User

	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	location, err := time.LoadLocation(user.Timezone)
	if err != nil {
		location = time.UTC
	}

	now := time.Now().In(location)

	today := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		location,
	)

	var session models.VocabularySession

	err = db.
		Where("user_id = ? AND date = ?", user.ID, today).
		First(&session).Error

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No vocabulary session found for today",
		})
		return
	}

	if session.XPRewarded {
		c.JSON(http.StatusOK, gin.H{
			"message": "Vocabulary already completed",
			"xp":      0,
		})
		return
	}

	const xpReward = 50

	user.XP += xpReward
	session.Completed = true
	session.XPRewarded = true

	db.Save(&user)
	db.Save(&session)

	c.JSON(http.StatusOK, gin.H{
		"message": "Vocabulary session completed",
		"xp":      xpReward,
	})
}
