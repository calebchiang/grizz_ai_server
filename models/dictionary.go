package models

import "time"

type Dictionary struct {
	ID uint `gorm:"primaryKey"`

	UserID uint `gorm:"not null;uniqueIndex:idx_user_vocab"`
	User   User `gorm:"foreignKey:UserID"`

	VocabularyID uint       `gorm:"not null;uniqueIndex:idx_user_vocab"`
	Vocabulary   Vocabulary `gorm:"foreignKey:VocabularyID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
