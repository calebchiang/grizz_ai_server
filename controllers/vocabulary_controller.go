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
	// GET USER ID FROM CONTEXT
	// ---------------------------

	userID := c.GetUint("userID")

	var user models.User

	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
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
	// CHECK FOR EXISTING SESSION
	// ---------------------------

	var session models.VocabularySession

	err = db.
		Where("user_id = ? AND date = ?", user.ID, today).
		First(&session).Error

	if err != nil {

		// ---------------------------
		// CREATE NEW SESSION
		// ---------------------------

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

	// Prevent null JSON
	if words == nil {
		words = []models.Vocabulary{}
	}

	// ---------------------------
	// RESPONSE
	// ---------------------------

	c.JSON(http.StatusOK, gin.H{
		"words":     words,
		"completed": session.Completed,
	})
}

func CompleteVocabularySession(c *gin.Context) {

	db := database.DB

	// ---------------------------
	// GET USER
	// ---------------------------

	userID := c.GetUint("userID")

	var user models.User

	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
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
	// FIND TODAY SESSION
	// ---------------------------

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

	// ---------------------------
	// PREVENT DOUBLE XP
	// ---------------------------

	if session.XPRewarded {
		c.JSON(http.StatusOK, gin.H{
			"message": "Vocabulary already completed",
			"xp":      0,
		})
		return
	}

	// ---------------------------
	// REWARD XP
	// ---------------------------

	const xpReward = 50

	user.XP += xpReward
	session.Completed = true
	session.XPRewarded = true

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user XP",
		})
		return
	}

	if err := db.Save(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update session",
		})
		return
	}

	// ---------------------------
	// RESPONSE
	// ---------------------------

	c.JSON(http.StatusOK, gin.H{
		"message": "Vocabulary session completed",
		"xp":      xpReward,
	})
}
