package models

import (
	"time"
)

// UserStory represents a high-level requirement or feature.
type UserStory struct {
	ID          string     `gorm:"primaryKey;type:string" json:"id"`
	ProjectID   string     `gorm:"not null" json:"projectId"`
	Title       string     `gorm:"not null" json:"title"`
	Description string     `gorm:"not null" json:"description"`
	Acceptance  *string    `json:"acceptance,omitempty"`
	Priority    string     `gorm:"default:'MEDIUM'" json:"priority"` // MEDIUM, HIGH, LOW
	StoryPoints *int       `json:"storyPoints,omitempty"`
	Status      string     `gorm:"default:'BACKLOG'" json:"status"` // BACKLOG, TODO, IN_PROGRESS, DONE
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	AssigneeID  *string
	Assignee    *User      `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	SprintID    *string
	Sprint      *Sprint    `gorm:"foreignKey:SprintID" json:"sprint,omitempty"`
	Project     Project    `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"-"`
	Tasks       []Task     `gorm:"foreignKey:UserStoryID" json:"tasks,omitempty"`
}

// TableName specifies the custom table name for UserStory.
func (UserStory) TableName() string {
	return "user_stories"
}

// Task represents a specific action item within a project or story.
type Task struct {
	ID          string     `gorm:"primaryKey;type:string" json:"id"`
	ProjectID   string     `gorm:"not null" json:"projectId"`
	UserStoryID *string    `json:"userStoryId,omitempty"`
	SprintID    *string    `json:"sprintID,omitempty"`
	Title       string     `gorm:"not null" json:"title"`
	Description *string    `json:"description,omitempty"`
	Priority    string     `gorm:"default:'MEDIUM'" json:"priority"`
	Status      string     `gorm:"default:'TODO'" json:"status"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	AssigneeID  *string
	Assignee    *User      `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Project     Project    `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"-"`
	UserStory   *UserStory `gorm:"foreignKey:UserStoryID" json:"userStory,omitempty"`
	Sprint      *Sprint    `gorm:"foreignKey:SprintID" json:"sprint,omitempty"`
	Evaluations []Evaluation `gorm:"foreignKey:TaskID" json:"evaluations,omitempty"`
}

// TableName specifies the custom table name for Task.
func (Task) TableName() string {
	return "tasks"
}
