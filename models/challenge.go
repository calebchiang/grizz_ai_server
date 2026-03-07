package models

import "time"

type Challenge struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"not null"`
	Description string `gorm:"type:text"`
	XPReward    int    `gorm:"not null;default:25"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
