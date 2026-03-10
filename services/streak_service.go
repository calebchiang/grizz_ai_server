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

	// ------------------------------------------------
	// If user has previous activity
	// ------------------------------------------------

	if user.LastActivityAt != nil {

		last := user.LastActivityAt.In(location)

		lastDay := time.Date(
			last.Year(),
			last.Month(),
			last.Day(),
			0, 0, 0, 0,
			location,
		)

		// Already counted today
		if lastDay.Equal(startOfToday) {
			return
		}

		// Continue streak if yesterday
		if lastDay.Equal(startOfYesterday) {
			user.CurrentStreak++
		} else {
			// Streak broken
			user.CurrentStreak = 1
		}

	} else {

		// First ever activity
		user.CurrentStreak = 1
	}

	// Update longest streak
	if user.CurrentStreak > user.LongestStreak {
		user.LongestStreak = user.CurrentStreak
	}

	// Update last activity time
	user.LastActivityAt = &now

	database.DB.Save(&user)
}
