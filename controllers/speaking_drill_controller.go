package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/calebchiang/thirdparty_server/services"
	"golang.org/x/sync/errgroup"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

func StartSpeakingDrill(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// -------- PARSE FORM INPUT --------

	topic := c.PostForm("topic")
	transcript := c.PostForm("transcript")

	if topic == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Topic is required",
		})
		return
	}

	if transcript == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Transcript is required",
		})
		return
	}

	// Optional video file
	videoFile, _ := c.FormFile("video")

	// -------- CONCURRENT TASKS --------

	var (
		result   *services.DrillResult
		videoURL *string
	)

	var g errgroup.Group

	// AI analysis
	g.Go(func() error {

		r, err := services.GenerateDrillResult(topic, transcript)
		if err != nil {
			return err
		}

		result = r
		return nil
	})

	// Video upload (optional)
	if videoFile != nil {

		g.Go(func() error {

			url, err := services.UploadVideoToR2(videoFile)
			if err != nil {
				return err
			}

			videoURL = &url
			return nil
		})
	}

	// Wait for both tasks
	if err := g.Wait(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process speaking drill",
		})
		return
	}

	// -------- CONVERT ARRAYS → JSONB --------

	fillerWordsJSON, _ := json.Marshal(result.FillerWords)
	strengthsJSON, _ := json.Marshal(result.Strengths)
	weaknessesJSON, _ := json.Marshal(result.Weaknesses)
	phrasesJSON, _ := json.Marshal(result.PhraseReplacements)

	// -------- CREATE DRILL --------

	drill := models.SpeakingDrill{
		UserID:     userID.(uint),
		Topic:      topic,
		Transcript: transcript,
		VideoURL:   videoURL,

		Clarity:      result.Scores.Clarity,
		Articulation: result.Scores.Articulation,
		FillerRate:   result.Scores.FillerRate,
		Pace:         result.Scores.Pace,
		Structure:    result.Scores.Structure,

		SpeakingScore: result.SpeakingScore,

		FillerWords:         datatypes.JSON(fillerWordsJSON),
		Strengths:           datatypes.JSON(strengthsJSON),
		Weaknesses:          datatypes.JSON(weaknessesJSON),
		PhrasesToUseInstead: datatypes.JSON(phrasesJSON),
	}

	if err := database.DB.Create(&drill).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create speaking drill",
		})
		return
	}

	// -------- RESPONSE --------

	c.JSON(http.StatusCreated, gin.H{
		"message": "Speaking drill created",

		"drill_id":   drill.ID,
		"topic":      drill.Topic,
		"created_at": drill.CreatedAt,
		"video_url":  drill.VideoURL,
		"transcript": drill.Transcript,

		// Scores
		"clarity":        drill.Clarity,
		"articulation":   drill.Articulation,
		"filler_rate":    drill.FillerRate,
		"pace":           drill.Pace,
		"structure":      drill.Structure,
		"speaking_score": drill.SpeakingScore,

		// AI feedback
		"filler_words":        result.FillerWords,
		"strengths":           result.Strengths,
		"weaknesses":          result.Weaknesses,
		"phrase_replacements": result.PhraseReplacements,
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
	completedDaysMap := map[int]bool{}

	for _, drill := range drills {

		local := drill.CreatedAt.In(loc)

		if local.Month() == now.Month() && local.Year() == now.Year() {
			completedDaysMap[local.Day()] = true
		}
	}

	completedDays := []int{}

	for day := range completedDaysMap {
		completedDays = append(completedDays, day)
	}

	c.JSON(http.StatusOK, gin.H{
		"current_day":    currentDay,
		"completed_days": completedDays,
	})
}
