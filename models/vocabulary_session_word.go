package models

import "time"

type VocabularySessionWord struct {
	ID uint `gorm:"primaryKey"`

	SessionID uint              `gorm:"index;not null"`
	Session   VocabularySession `gorm:"foreignKey:SessionID"`

	VocabularyID uint       `gorm:"index;not null"`
	Vocabulary   Vocabulary `gorm:"foreignKey:VocabularyID"`

	OrderIndex int  `gorm:"not null"` // preserves the 1-5 order shown to the user
	Completed  bool `gorm:"default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
