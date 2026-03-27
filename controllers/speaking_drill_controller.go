package controllers

import (
	"net/http"
	"time"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/gin-gonic/gin"
)

func StartSpeakingDrill(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var input struct {
		Topic      string `json:"topic"`
		Transcript string `json:"transcript"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if input.Topic == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Topic is required",
		})
		return
	}

	if input.Transcript == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Transcript is required",
		})
		return
	}

	drill := models.SpeakingDrill{
		UserID:     userID.(uint),
		Topic:      input.Topic,
		Transcript: input.Transcript,
	}

	if err := database.DB.Create(&drill).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create speaking drill",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Speaking drill created",
		"drill_id":   drill.ID,
		"topic":      drill.Topic,
		"transcript": drill.Transcript,
		"created_at": drill.CreatedAt,
	})
}

func GetChallengeStatus(c *gin.Context) {

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userID := userIDRaw.(uint)

	// Get user to access timezone
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load user",
		})
		return
	}

	// Load timezone
	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		loc = time.UTC
	}

	// Fetch recent drills (40 days safe window)
	var drills []models.SpeakingDrill
	if err := database.DB.
		Where("user_id = ?", userID).
		Where("created_at >= NOW() - INTERVAL '40 days'").
		Find(&drills).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch drills",
		})
		return
	}

	// Map of completed days
	completed := map[string]bool{}

	for _, drill := range drills {

		local := drill.CreatedAt.In(loc)
		day := local.Format("2006-01-02")

		completed[day] = true
	}

	// Determine streak ending yesterday
	now := time.Now().In(loc)
	yesterday := now.AddDate(0, 0, -1)

	streak := 0

	for i := 0; i < 30; i++ {

		d := yesterday.AddDate(0, 0, -i)
		key := d.Format("2006-01-02")

		if completed[key] {
			streak++
		} else {
			break
		}
	}

	currentDay := streak + 1

	// Completed days this month
	var completedDays []int

	for day := range completed {

		parsed, err := time.Parse("2006-01-02", day)
		if err != nil {
			continue
		}

		parsed = parsed.In(loc)

		if parsed.Month() == now.Month() && parsed.Year() == now.Year() {
			completedDays = append(completedDays, parsed.Day())
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"current_day":    currentDay,
		"completed_days": completedDays,
	})
}
