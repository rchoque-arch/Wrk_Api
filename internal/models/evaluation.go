package models

import (
	"time"
)

type Rubric struct {
	ID          string     `gorm:"primaryKey;type:string" json:"id"`
	ProjectID   *string    `json:"projectId,omitempty"`
	Name        string     `gorm:"not null" json:"name"`
	Description *string    `json:"description,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	Project     *Project   `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"-"`
	Criteria    []Criteria `gorm:"foreignKey:RubricID" json:"criteria,omitempty"`
}

func (Rubric) TableName() string {
	return "rubrics"
}

type Criteria struct {
	ID          string     `gorm:"primaryKey;type:string" json:"id"`
	RubricID    string     `gorm:"not null" json:"rubricId"`
	Name        string     `gorm:"not null" json:"name"`
	Description *string    `json:"description,omitempty"`
	MaxScore    int        `gorm:"default:100" json:"maxScore"`
	Weight      int        `gorm:"default:1" json:"weight"`

	Rubric      Rubric     `gorm:"foreignKey:RubricID;constraint:OnDelete:CASCADE" json:"-"`
}

func (Criteria) TableName() string {
	return "criteria"
}

type Evaluation struct {
	ID          string     `gorm:"primaryKey;type:string" json:"id"`
	ProjectID   string     `gorm:"not null" json:"projectId"`
	TaskID      *string    `json:"taskId,omitempty"`
	SprintID    *string    `json:"sprintId,omitempty"`
	EvaluatorID string     `gorm:"not null" json:"evaluatorId"`
	Status      string     `gorm:"default:'PENDING'" json:"status"` // PENDING, COMPLETED
	Feedback    *string    `json:"feedback,omitempty"`
	Score       *int       `json:"score,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relationships
	Project     Project     `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"-"`
	Task        *Task       `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"task,omitempty"`
	Sprint      *Sprint     `gorm:"foreignKey:SprintID;constraint:OnDelete:CASCADE" json:"sprint,omitempty"`
	Evaluator   User        `gorm:"foreignKey:EvaluatorID" json:"evaluator,omitempty"`
	Criteria    []EvaluationCriteria `gorm:"foreignKey:EvaluationID" json:"criteria,omitempty"`
}

func (Evaluation) TableName() string {
	return "evaluations"
}

type EvaluationCriteria struct {
	ID           string     `gorm:"primaryKey;type:string" json:"id"`
	EvaluationID string     `gorm:"not null;uniqueIndex:idx_eval_crit" json:"evaluationId"`
	CriteriaID   string     `gorm:"not null;uniqueIndex:idx_eval_crit" json:"criteriaId"`
	Score        int        `gorm:"default:0" json:"score"`
	Comment      *string    `json:"comment,omitempty"`

	Evaluation   Evaluation `gorm:"foreignKey:EvaluationID;constraint:OnDelete:CASCADE" json:"-"`
	Criteria     Criteria   `gorm:"foreignKey:CriteriaID" json:"criteria,omitempty"`
}

func (EvaluationCriteria) TableName() string {
	return "evaluation_criteria"
}
