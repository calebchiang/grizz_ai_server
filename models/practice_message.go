package models

import "time"

type PracticeMessage struct {
	ID        uint   `gorm:"primaryKey"`
	SessionID uint   `gorm:"not null;index"`
	Role      string `gorm:"not null"` // "user" or "assistant"
	Content   string `gorm:"type:text;not null"`
	CreatedAt time.Time

	Session PracticeSession `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE"`
}
