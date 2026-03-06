package models

import (
	"time"
)

// User represents an account in the system.
type User struct {
	ID        string    `gorm:"primaryKey;type:string" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Name      string    `gorm:"not null" json:"name"`
	Password  string    `gorm:"not null" json:"-"` // Don't return password in JSON
	Role      string    `gorm:"default:'TEAM_DEVELOPER'" json:"role"`
	Avatar    *string   `json:"avatar,omitempty"`
	Active    bool      `gorm:"default:true" json:"active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships - usually omitted in list views or circular references
	Projects           []Project           `gorm:"foreignKey:OwnerID" json:"projects,omitempty"`
	Tasks              []Task              `gorm:"foreignKey:AssigneeID" json:"tasks,omitempty"`
	Evaluations        []Evaluation        `gorm:"foreignKey:EvaluatorID" json:"evaluations,omitempty"`
	UserStories        []UserStory         `gorm:"foreignKey:AssigneeID" json:"userStories,omitempty"`
	Messages           []Message           `gorm:"foreignKey:UserID" json:"-"`
	ChatParticipants   []ChatParticipant   `gorm:"foreignKey:UserID" json:"-"`
	ProjectMemberships []ProjectMember     `gorm:"foreignKey:UserID" json:"-"`
	Notifications      []Notification      `gorm:"foreignKey:UserID" json:"-"`
	RetrospectiveItems []RetrospectiveItem `gorm:"foreignKey:UserID" json:"-"`
}

// TableName specifies the custom table name for User.
func (User) TableName() string {
	return "users"
}
