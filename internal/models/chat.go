// Package models provides models functionality.
package models

import (
	"time"
)

// Chat represents the Chat structure.
type Chat struct {
	ID        string `gorm:"primaryKey;type:string"`
	ProjectID *string
	Title     *string
	Type      string    `gorm:"default:'PROJECT'"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Relationships
	Project      *Project          `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Messages     []Message         `gorm:"foreignKey:ChatID"`
	Participants []ChatParticipant `gorm:"foreignKey:ChatID"`
}

// TableName overrides the table name used by GORM for t.
func (Chat) TableName() string {
	return "chats"
}

// ChatParticipant represents the ChatParticipant structure.
type ChatParticipant struct {
	ChatID string `gorm:"primaryKey;type:string"`
	UserID string `gorm:"primaryKey;type:string"`

	Chat Chat `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE"`
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName overrides the table name used by GORM for t.
func (ChatParticipant) TableName() string {
	return "chat_participants"
}

// Message represents the Message structure.
type Message struct {
	ID        string    `gorm:"primaryKey;type:string"`
	ChatID    string    `gorm:"not null"`
	UserID    string    `gorm:"not null"`
	Content   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	Chat Chat `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE"`
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName overrides the table name used by GORM for e.
func (Message) TableName() string {
	return "messages"
}
