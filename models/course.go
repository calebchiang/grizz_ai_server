package models

import "time"

type Course struct {
	ID uint `gorm:"primaryKey"`

	Category    string `gorm:"not null;index"` // Public Speaking, Social Skills, etc.
	Title       string `gorm:"not null"`
	Description string `gorm:"type:text"`

	SortOrder   int  `gorm:"not null;default:0"`
	IsPublished bool `gorm:"not null;default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Lessons []Lesson `gorm:"foreignKey:CourseID"`
}
