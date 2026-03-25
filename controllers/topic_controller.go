package controllers

import (
	"net/http"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/gin-gonic/gin"
)

func GetTopics(c *gin.Context) {

	var topics []models.Topic

	if err := database.DB.
		Order("RANDOM()").
		Limit(5).
		Find(&topics).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch topics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"topics": topics,
	})
}
