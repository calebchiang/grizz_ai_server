package models

import "time"

type PracticeSession struct {
	ID              uint      `gorm:"primaryKey"`
	UserID          uint      `gorm:"not null;index"` // FK -> users.id
	Scenario        string    `gorm:"not null"`
	Persona         string    `gorm:"not null"`
	DurationSeconds int       `gorm:"not null;default:0"`
	StartedAt       time.Time `gorm:"not null"`
	EndedAt         *time.Time
	Transcript      string `gorm:"type:text"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	// Optional (nice to have for preload later)
	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
