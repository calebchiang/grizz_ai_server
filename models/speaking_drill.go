package models

import (
	"time"

	"gorm.io/datatypes"
)

type SpeakingDrill struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint `gorm:"not null;index"`

	Topic      string `gorm:"not null"`
	Transcript string `gorm:"type:text"`

	// Speaking delivery traits (0–10 each)
	Clarity      int `gorm:"default:0"`
	Articulation int `gorm:"default:0"`
	FillerRate   int `gorm:"default:0"` // frequency of filler words
	Pace         int `gorm:"default:0"`
	Structure    int `gorm:"default:0"`

	// Final calculated speaking score (0–100)
	SpeakingScore int `gorm:"default:0"`

	FillerWords datatypes.JSON `gorm:"type:jsonb"`

	// AI feedback
	Strengths           datatypes.JSON `gorm:"type:jsonb"`
	Weaknesses          datatypes.JSON `gorm:"type:jsonb"`
	PhrasesToUseInstead datatypes.JSON `gorm:"type:jsonb"`

	CreatedAt time.Time
	UpdatedAt time.Time

	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
