package services

import (
	"time"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
)

func UpdateUserStreak(userID uint) {

	var user models.User

	if err := database.DB.First(&user, userID).Error; err != nil {
		return
	}

	// Load user timezone
	location, err := time.LoadLocation(user.Timezone)
	if err != nil {
		location = time.UTC
	}

	now := time.Now().In(location)

	startOfToday := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		location,
	)

	startOfYesterday := startOfToday.Add(-24 * time.Hour)

	endOfToday := startOfToday.Add(24 * time.Hour)

	// ------------------------------------------------
	// Check if activity already existed today BEFORE this event
	// (meaning streak already counted)
	// ------------------------------------------------

	var practiceTodayCount int64

	database.DB.Model(&models.PracticeSession{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startOfToday, endOfToday).
		Count(&practiceTodayCount)

	var challengeTodayCount int64

	database.DB.Model(&models.ChallengeCompletion{}).
		Where("user_id = ? AND date >= ? AND date < ?", userID, startOfToday, endOfToday).
		Count(&challengeTodayCount)

	// If more than 1 activity exists today, streak was already incremented
	if practiceTodayCount+challengeTodayCount > 1 {
		return
	}

	// ------------------------------------------------
	// Check if user had activity yesterday
	// ------------------------------------------------

	var practiceYesterdayCount int64

	database.DB.Model(&models.PracticeSession{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startOfYesterday, startOfToday).
		Count(&practiceYesterdayCount)

	var challengeYesterdayCount int64

	database.DB.Model(&models.ChallengeCompletion{}).
		Where("user_id = ? AND date >= ? AND date < ?", userID, startOfYesterday, startOfToday).
		Count(&challengeYesterdayCount)

	hadActivityYesterday := practiceYesterdayCount+challengeYesterdayCount > 0

	// ------------------------------------------------
	// Update streak
	// ------------------------------------------------

	if hadActivityYesterday {
		user.CurrentStreak++
	} else {
		user.CurrentStreak = 1
	}

	if user.CurrentStreak > user.LongestStreak {
		user.LongestStreak = user.CurrentStreak
	}

	database.DB.Save(&user)
}
