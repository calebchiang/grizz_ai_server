package controllers

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/gin-gonic/gin"
)

func GetTodayChallenges(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get user
	var user models.User
	if err := database.DB.First(&user, userID.(uint)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Load timezone
	location, err := time.LoadLocation(user.Timezone)
	if err != nil {
		location = time.UTC
	}

	now := time.Now().In(location)

	// Daily seed
	dateKey := now.Format("2006-01-02")

	seed := int64(user.ID)
	for _, c := range dateKey {
		seed += int64(c)
	}

	r := rand.New(rand.NewSource(seed))

	// Get all challenges
	var challenges []models.Challenge
	if err := database.DB.Find(&challenges).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load challenges"})
		return
	}

	if len(challenges) < 3 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Not enough challenges"})
		return
	}

	// Shuffle deterministically
	r.Shuffle(len(challenges), func(i, j int) {
		challenges[i], challenges[j] = challenges[j], challenges[i]
	})

	selected := challenges[:3]

	c.JSON(http.StatusOK, selected)
}
