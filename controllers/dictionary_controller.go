package controllers

import (
	"net/http"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/gin-gonic/gin"
)

func SaveWord(c *gin.Context) {

	// -------------------------
	// GET USER ID FROM JWT
	// -------------------------

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// -------------------------
	// PARSE REQUEST BODY
	// -------------------------

	var input struct {
		VocabularyID uint `json:"vocabulary_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if input.VocabularyID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Vocabulary ID required",
		})
		return
	}

	// -------------------------
	// CREATE DICTIONARY ENTRY
	// -------------------------

	entry := models.Dictionary{
		UserID:       userID.(uint),
		VocabularyID: input.VocabularyID,
	}

	if err := database.DB.Create(&entry).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save word",
		})
		return
	}

	// -------------------------
	// RESPONSE
	// -------------------------

	c.JSON(http.StatusOK, gin.H{
		"message": "Word saved",
	})
}
