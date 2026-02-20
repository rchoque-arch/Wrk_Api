package models

import (
	"time"
)

type Notification struct {
	ID        string    `gorm:"primaryKey;type:string"`
	UserID    string    `gorm:"not null"`
	Title     string    `gorm:"not null"`
	Message   string    `gorm:"not null"`
	Type      string    `gorm:"not null"` // TASK_ASSIGNED, EVALUATION_COMPLETED, MESSAGE
	Read      bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (Notification) TableName() string {
	return "notifications"
}
