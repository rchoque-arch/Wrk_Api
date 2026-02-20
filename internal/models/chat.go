package models

import (
	"time"
)

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

func (Chat) TableName() string {
	return "chats"
}

type ChatParticipant struct {
	ChatID string `gorm:"primaryKey;type:string"`
	UserID string `gorm:"primaryKey;type:string"`

	Chat Chat `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE"`
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (ChatParticipant) TableName() string {
	return "chat_participants"
}

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

func (Message) TableName() string {
	return "messages"
}
