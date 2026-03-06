package models

import (
	"time"
)

// Project represents a workspace for tasks and documents.
type Project struct {
	ID          string     `gorm:"primaryKey;type:string" json:"id"`
	Name        string     `gorm:"not null" json:"name"`
	Description *string    `json:"description,omitempty"`
	Status      string     `gorm:"default:'ACTIVE'" json:"status"`
	StartDate   *time.Time `json:"startDate,omitempty"`
	EndDate     *time.Time `json:"endDate,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	OwnerID     string
	Owner       User            `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE" json:"owner,omitempty"`
	Members     []ProjectMember `gorm:"foreignKey:ProjectID" json:"members,omitempty"`
	Sprints     []Sprint        `gorm:"foreignKey:ProjectID" json:"sprints,omitempty"`
	UserStories []UserStory     `gorm:"foreignKey:ProjectID" json:"userStories,omitempty"`
	Tasks       []Task          `gorm:"foreignKey:ProjectID" json:"tasks,omitempty"`
	Evaluations []Evaluation    `gorm:"foreignKey:ProjectID" json:"evaluations,omitempty"`
	Rubrics     []Rubric        `gorm:"foreignKey:ProjectID" json:"rubrics,omitempty"`
	Chats       []Chat          `gorm:"foreignKey:ProjectID" json:"chats,omitempty"`
	Documents   []Document      `gorm:"foreignKey:ProjectID" json:"documents,omitempty"`
}

// TableName specifies the custom table name for Project.
func (Project) TableName() string {
	return "projects"
}

// ProjectMember represents a user participating in a project.
type ProjectMember struct {
	ID        string    `gorm:"primaryKey;type:string" json:"id"`
	ProjectID string    `gorm:"not null;uniqueIndex:idx_project_member" json:"projectId"`
	UserID    string    `gorm:"not null;uniqueIndex:idx_project_member" json:"userId"`
	Role      string    `gorm:"not null" json:"role"`
	JoinedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"joinedAt"`

	Project Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
	User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName specifies the custom table name for ProjectMember.
func (ProjectMember) TableName() string {
	return "project_members"
}

// Document represents a file or text attached to a project.
type Document struct {
	ID        string    `gorm:"primaryKey;type:string" json:"id"`
	ProjectID string    `gorm:"not null" json:"projectId"`
	Name      string    `gorm:"not null" json:"name"`
	URL       string    `gorm:"not null" json:"url"` // Simulated
	Type      string    `gorm:"not null" json:"type"` // PDF, DOCX, etc.
	Size      *int      `json:"size,omitempty"`       // KB
	Version   int       `gorm:"default:1" json:"version"`
	ParentID  *string   `json:"parentId,omitempty"`
	UploadedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"uploadedAt"`

	Project   Project   `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project,omitempty"`
	Parent    *Document `gorm:"foreignKey:ParentID;constraint:OnDelete:SET NULL" json:"parent,omitempty"`
	Versions  []Document `gorm:"foreignKey:ParentID" json:"versions,omitempty"`
}

// TableName specifies the custom table name for Document.
func (Document) TableName() string {
	return "documents"
}
