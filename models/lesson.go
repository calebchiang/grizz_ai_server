package models

import "time"

type Lesson struct {
	ID uint `gorm:"primaryKey"`

	CourseID uint `gorm:"not null;index"`

	Title string `gorm:"not null"`
	Goal  string `gorm:"type:text"`

	EstimatedMinutes int `gorm:"not null;default:3"`
	XPReward         int `gorm:"not null;default:25"`

	SortOrder   int  `gorm:"not null;default:0"`
	IsPublished bool `gorm:"not null;default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Course Course `gorm:"foreignKey:CourseID"`

	Blocks []LessonBlock `gorm:"foreignKey:LessonID"`
}
