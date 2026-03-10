package controllers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CreateUser(c *gin.Context) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Basic validation
	if input.Email == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password required"})
		return
	}

	// Normalize email
	email := strings.ToLower(strings.TrimSpace(input.Email))

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Name:     input.Name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
	})
}

func LoginUser(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if input.Email == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password required"})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT secret not configured"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
	})
}

func GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var user models.User

	if err := database.DB.
		Select("id, name, email, credits, xp, current_streak, longest_streak").
		Where("id = ?", userID.(uint)).
		First(&user).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"name":           user.Name,
		"email":          user.Email,
		"credits":        user.Credits,
		"xp":             user.XP,
		"current_streak": user.CurrentStreak,
		"longest_streak": user.LongestStreak,
	})
}

func UpdateUserName(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var input struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	name := strings.TrimSpace(input.Name)

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name cannot be empty",
		})
		return
	}

	if err := database.DB.Model(&models.User{}).
		Where("id = ?", userID.(uint)).
		Update("name", name).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update name",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name": name,
	})
}

func AddXP(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	const reward = 25

	if err := database.DB.Model(&models.User{}).
		Where("id = ?", userID.(uint)).
		UpdateColumn("xp", gorm.Expr("xp + ?", reward)).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add XP",
		})
		return
	}

	var user models.User
	if err := database.DB.
		Select("xp").
		Where("id = ?", userID.(uint)).
		First(&user).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch XP",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"xp": user.XP,
	})
}

func GetWeeklyOverview(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

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

	// Start of week (Sunday)
	weekday := int(now.Weekday())

	weekStart := time.Date(
		now.Year(),
		now.Month(),
		now.Day()-weekday,
		0, 0, 0, 0,
		location,
	)

	type DayStatus struct {
		Day    string `json:"day"`
		Date   string `json:"date"`
		Status string `json:"status"`
	}

	var result []DayStatus

	for i := 0; i < 7; i++ {

		day := weekStart.AddDate(0, 0, i)

		startOfDay := time.Date(
			day.Year(),
			day.Month(),
			day.Day(),
			0, 0, 0, 0,
			location,
		)

		endOfDay := startOfDay.Add(24 * time.Hour)

		var practiceCount int64
		var challengeCount int64

		database.DB.Model(&models.PracticeSession{}).
			Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startOfDay, endOfDay).
			Count(&practiceCount)

		// Check challenges
		database.DB.Model(&models.ChallengeCompletion{}).
			Where("user_id = ? AND date >= ? AND date < ?", userID, startOfDay, endOfDay).
			Count(&challengeCount)

		completed := practiceCount > 0 || challengeCount > 0

		var status string

		if day.After(now) {

			status = "future"

		} else if day.Year() == now.Year() &&
			day.YearDay() == now.YearDay() {

			if completed {
				status = "completed"
			} else {
				status = "pending"
			}

		} else {

			if completed {
				status = "completed"
			} else {
				status = "uncompleted"
			}
		}

		result = append(result, DayStatus{
			Day:    day.Weekday().String(),
			Date:   day.Format("2006-01-02"),
			Status: status,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"week": result,
	})
}

func GetRecentPractice(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

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

	type DayStatus struct {
		Label  string `json:"label"`
		Date   string `json:"date"`
		Status string `json:"status"`
	}

	var result []DayStatus

	labels := []string{"today", "yesterday", "two_days_ago"}

	for i := 0; i < 3; i++ {

		day := now.AddDate(0, 0, -i)

		startOfDay := time.Date(
			day.Year(),
			day.Month(),
			day.Day(),
			0, 0, 0, 0,
			location,
		)

		endOfDay := startOfDay.Add(24 * time.Hour)

		var practiceCount int64

		database.DB.Model(&models.PracticeSession{}).
			Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startOfDay, endOfDay).
			Count(&practiceCount)

		completed := practiceCount > 0

		var status string

		if i == 0 { // today

			if completed {
				status = "completed"
			} else {
				status = "pending"
			}

		} else {

			if completed {
				status = "completed"
			} else {
				status = "uncompleted"
			}
		}

		result = append(result, DayStatus{
			Label:  labels[i],
			Date:   startOfDay.Format("2006-01-02"),
			Status: status,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"recent": result,
	})
}

func GetPracticeChallengeOverview(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// -------- TOTAL PRACTICE (ALL TIME) --------

	var totalPractice int64
	database.DB.Model(&models.PracticeSession{}).
		Where("user_id = ?", userID).
		Count(&totalPractice)

	// -------- TOTAL CHALLENGES (ALL TIME) --------

	var totalChallenges int64
	database.DB.Model(&models.ChallengeCompletion{}).
		Where("user_id = ?", userID).
		Count(&totalChallenges)

	// -------- HIGHEST CONVERSATION SCORE --------

	var highestScore int
	database.DB.Model(&models.PracticeSession{}).
		Where("user_id = ?", userID).
		Select("COALESCE(MAX(conversation_score), 0)").
		Scan(&highestScore)

	// -------- DAYS PRACTICED THIS WEEK --------

	var daysPracticedThisWeek int64
	database.DB.Raw(`
		SELECT COUNT(DISTINCT DATE(created_at))
		FROM practice_sessions
		WHERE user_id = ?
		AND created_at >= DATE_TRUNC('week', NOW())
	`, userID).Scan(&daysPracticedThisWeek)

	// -------- DAYS PRACTICED LAST WEEK --------

	var daysPracticedLastWeek int64
	database.DB.Raw(`
		SELECT COUNT(DISTINCT DATE(created_at))
		FROM practice_sessions
		WHERE user_id = ?
		AND created_at >= DATE_TRUNC('week', NOW()) - INTERVAL '1 week'
		AND created_at < DATE_TRUNC('week', NOW())
	`, userID).Scan(&daysPracticedLastWeek)

	// -------- WEEKLY PRACTICE (LAST 4 WEEKS) --------

	type WeeklyCount struct {
		Week  time.Time `json:"week"`
		Count int       `json:"count"`
	}

	var weeklyPractice []WeeklyCount

	database.DB.Raw(`
		SELECT DATE_TRUNC('week', created_at) AS week, COUNT(*) AS count
		FROM practice_sessions
		WHERE user_id = ?
		AND created_at >= NOW() - INTERVAL '4 weeks'
		GROUP BY week
		ORDER BY week
	`, userID).Scan(&weeklyPractice)

	// -------- WEEKLY CHALLENGES (LAST 4 WEEKS) --------

	var weeklyChallenges []WeeklyCount

	database.DB.Raw(`
		SELECT DATE_TRUNC('week', date) AS week, COUNT(*) AS count
		FROM challenge_completions
		WHERE user_id = ?
		AND date >= NOW() - INTERVAL '4 weeks'
		GROUP BY week
		ORDER BY week
	`, userID).Scan(&weeklyChallenges)

	// -------- FIX: Prevent null arrays in JSON --------

	if weeklyPractice == nil {
		weeklyPractice = []WeeklyCount{}
	}

	if weeklyChallenges == nil {
		weeklyChallenges = []WeeklyCount{}
	}

	// -------- RESPONSE --------

	c.JSON(http.StatusOK, gin.H{
		"total_practice":             totalPractice,
		"total_challenges_completed": totalChallenges,

		"highest_conversation_score": highestScore,

		"days_practiced_this_week": daysPracticedThisWeek,
		"days_practiced_last_week": daysPracticedLastWeek,

		"weekly_practice_counts":  weeklyPractice,
		"weekly_challenge_counts": weeklyChallenges,
	})
}
