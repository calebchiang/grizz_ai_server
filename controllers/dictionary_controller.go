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

func GetUserDictionary(c *gin.Context) {

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
	// FETCH USER DICTIONARY
	// -------------------------

	var dictionaryEntries []models.Dictionary

	err := database.DB.
		Preload("Vocabulary").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&dictionaryEntries).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch dictionary",
		})
		return
	}

	// -------------------------
	// EXTRACT VOCABULARY WORDS
	// -------------------------

	var words []models.Vocabulary

	for _, entry := range dictionaryEntries {
		words = append(words, entry.Vocabulary)
	}

	if words == nil {
		words = []models.Vocabulary{}
	}

	// -------------------------
	// RESPONSE
	// -------------------------

	c.JSON(http.StatusOK, gin.H{
		"words": words,
	})
}

func GetRecentWords(c *gin.Context) {

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
	// FETCH MOST RECENT WORDS
	// -------------------------

	var dictionaryEntries []models.Dictionary

	err := database.DB.
		Preload("Vocabulary").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(6).
		Find(&dictionaryEntries).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch recent words",
		})
		return
	}

	// -------------------------
	// EXTRACT VOCABULARY WORDS
	// -------------------------

	var words []models.Vocabulary

	for _, entry := range dictionaryEntries {
		words = append(words, entry.Vocabulary)
	}

	if words == nil {
		words = []models.Vocabulary{}
	}

	// -------------------------
	// RESPONSE
	// -------------------------

	c.JSON(http.StatusOK, gin.H{
		"words": words,
	})
}

func RemoveWord(c *gin.Context) {

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
	// GET VOCABULARY ID FROM PARAM
	// -------------------------

	vocabularyID := c.Param("vocabulary_id")

	if vocabularyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Vocabulary ID required",
		})
		return
	}

	// -------------------------
	// DELETE DICTIONARY ENTRY
	// -------------------------

	err := database.DB.
		Where("user_id = ? AND vocabulary_id = ?", userID, vocabularyID).
		Delete(&models.Dictionary{}).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove word",
		})
		return
	}

	// -------------------------
	// RESPONSE
	// -------------------------

	c.JSON(http.StatusOK, gin.H{
		"message": "Word removed from dictionary",
	})
}
