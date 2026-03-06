package models

import (
	"time"
)

// Chat represents a chat room or direct message conversation.
type Chat struct {
	ID        string     `gorm:"primaryKey;type:string"`
	ProjectID *string
	Title     *string
	Type      string     `gorm:"default:'PROJECT'"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`

	// Relationships
	Project      *Project          `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Messages     []Message         `gorm:"foreignKey:ChatID"`
	Participants []ChatParticipant `gorm:"foreignKey:ChatID"`
}

// TableName specifies the custom table name for Chat.
func (Chat) TableName() string {
	return "chats"
}

// ChatParticipant represents a user participating in a chat.
type ChatParticipant struct {
	ChatID string `gorm:"primaryKey;type:string"`
	UserID string `gorm:"primaryKey;type:string"`

	Chat Chat `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE"`
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the custom table name for ChatParticipant.
func (ChatParticipant) TableName() string {
	return "chat_participants"
}

// Message represents a single message in a chat.
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

// TableName specifies the custom table name for Message.
func (Message) TableName() string {
	return "messages"
}
