package models

import "time"

type Vocabulary struct {
	ID            uint   `gorm:"primaryKey"`
	Word          string `gorm:"uniqueIndex;not null"`
	Definition    string `gorm:"type:text;not null"`
	Pronunciation string `gorm:"not null"`
	PartOfSpeech  string `gorm:"type:varchar(20);not null"` // noun, verb, adjective, etc

	CreatedAt time.Time
	UpdatedAt time.Time
}
