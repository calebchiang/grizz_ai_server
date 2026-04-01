package controllers

import (
	"net/http"
	"strconv"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/gin-gonic/gin"
)

func GetRandomVocabulary(c *gin.Context) {

	// ---------------------------
	// READ LIMIT QUERY PARAM
	// ---------------------------

	limit := 1

	if query := c.Query("limit"); query != "" {
		if parsed, err := strconv.Atoi(query); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// ---------------------------
	// FETCH RANDOM WORDS
	// ---------------------------

	var words []models.Vocabulary

	err := database.DB.
		Order("RANDOM()").
		Limit(limit).
		Find(&words).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch vocabulary",
		})
		return
	}

	// Prevent null JSON
	if words == nil {
		words = []models.Vocabulary{}
	}

	// ---------------------------
	// RESPONSE
	// ---------------------------

	c.JSON(http.StatusOK, gin.H{
		"words": words,
	})
}
