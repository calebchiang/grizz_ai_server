package models

import "time"

type VocabularySession struct {
	ID uint `gorm:"primaryKey"`

	UserID uint `gorm:"index;not null"`
	User   User `gorm:"foreignKey:UserID"`

	Date time.Time `gorm:"index;not null"` // normalized to the user's local day

	Completed  bool `gorm:"not null;default:false"`
	XPRewarded bool `gorm:"not null;default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
