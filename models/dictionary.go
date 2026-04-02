package models

import "time"

type Dictionary struct {
	ID uint `gorm:"primaryKey"`

	UserID uint `gorm:"index;not null"`
	User   User `gorm:"foreignKey:UserID"`

	VocabularyID uint       `gorm:"index;not null"`
	Vocabulary   Vocabulary `gorm:"foreignKey:VocabularyID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
