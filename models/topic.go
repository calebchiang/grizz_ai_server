package models

import "time"

type Topic struct {
	ID        uint   `gorm:"primaryKey"`
	Topic     string `gorm:"not null;uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
