package models

import "time"

type User struct {
	ID                   uint       `gorm:"primaryKey"`
	Name                 string     `gorm:"not null"`
	Email                string     `gorm:"uniqueIndex;not null"`
	Password             string     `gorm:"not null"`
	Credits              int        `gorm:"not null;default:1"`
	XP                   int        `gorm:"not null;default:0"`
	IsPremium            bool       `gorm:"not null;default:false"`
	SeenOnboarding       bool       `gorm:"not null;default:false"`
	SeenAIDataDisclosure bool       `gorm:"not null;default:false"`
	Timezone             string     `gorm:"default:'UTC'"`
	CurrentStreak        int        `gorm:"not null;default:0"`
	LongestStreak        int        `gorm:"not null;default:0"`
	HeardFrom            string     `gorm:"type:varchar(50)"`
	AgeGroup             string     `gorm:"type:varchar(20)"`
	LastActivityAt       *time.Time `gorm:"index"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
