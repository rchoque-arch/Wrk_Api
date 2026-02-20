package models

import (
	"time"
)

type Sprint struct {
	ID          string     `gorm:"primaryKey;type:string" json:"id"`
	ProjectID   string     `gorm:"not null" json:"projectId"`
	Name        string     `gorm:"not null" json:"name"`
	Description *string    `json:"description,omitempty"`
	StartDate   time.Time  `gorm:"not null" json:"startDate"`
	EndDate     time.Time  `gorm:"not null" json:"endDate"`
	Status      string     `gorm:"default:'PLANNING'" json:"status"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	Project            Project             `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"-"`
	UserStories        []UserStory         `gorm:"foreignKey:SprintID" json:"userStories,omitempty"`
	Tasks              []Task              `gorm:"foreignKey:SprintID" json:"tasks,omitempty"`
	RetrospectiveItems []RetrospectiveItem `gorm:"foreignKey:SprintID" json:"retrospectiveItems,omitempty"`
	Evaluations        []Evaluation        `gorm:"foreignKey:SprintID" json:"evaluations,omitempty"`
}

func (Sprint) TableName() string {
	return "sprints"
}

type RetrospectiveItem struct {
	ID        string    `gorm:"primaryKey;type:string" json:"id"`
	SprintID  string    `gorm:"not null" json:"sprintId"`
	Type      string    `gorm:"not null" json:"type"` // GOOD, BAD, ACTION
	Content   string    `gorm:"not null" json:"content"`
	UserID    string    `gorm:"not null" json:"userId"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`

	Sprint Sprint `gorm:"foreignKey:SprintID;constraint:OnDelete:CASCADE" json:"-"`
	User   User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

func (RetrospectiveItem) TableName() string {
	return "retrospective_items"
}
