package models

import (
	"time"

	"gorm.io/datatypes"
)

type LessonBlock struct {
	ID uint `gorm:"primaryKey"`

	LessonID uint `gorm:"not null;index"`

	Type string `gorm:"not null"` // reflective, flashcard, mcq, fill_blank, summary, practice

	Title string `gorm:"type:text"`
	Body  string `gorm:"type:text"`

	// Used for MCQ / Fill in Blank / other structured blocks
	Options datatypes.JSON `gorm:"type:jsonb"`

	CorrectAnswer string `gorm:"type:text"`
	Explanation   string `gorm:"type:text"`

	SortOrder int `gorm:"not null;default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Lesson Lesson `gorm:"foreignKey:LessonID"`
}
