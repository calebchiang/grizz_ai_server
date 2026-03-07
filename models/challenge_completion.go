package models

import "time"

type ChallengeCompletion struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"not null;index"`
	ChallengeID uint      `gorm:"not null;index"`
	Date        time.Time `gorm:"not null;index"`

	CreatedAt time.Time
}
