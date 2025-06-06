package model

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Title        string       `json:"title" binding:"required"`
	Description  string       `json:"description"`
	Status       string       `gorm:"default:'Pending'" json:"status"`
	Priority     string       `gorm:"default:'Low'" json:"priority"`
	DueDate      time.Time    `json:"due_date"`
	CreatorID    uint         `json:"creator_id"`
	TeamID       uint         `json:"team_id" binding:"required"`
	Assignees    []User       `gorm:"many2many:task_assignees;constraint:OnDelete:CASCADE" json:"assignees"`
	ParentTaskID *uint        `json:"parent_task_id"`
	SubTasks     []Task       `gorm:"foreignKey:ParentTaskID;constraint:OnDelete:CASCADE" json:"subtasks"`
	Labels       []Label      `gorm:"many2many:task_labels;constraint:OnDelete:CASCADE" json:"labels"`
	Comments     []Comment    `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"comments"`
	Attachments  []Attachment `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"attachments"`
}

type CreateTaskInput struct {
	Title        string    `json:"title" binding:"required"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	Priority     string    `json:"priority"`
	DueDate      time.Time `json:"due_date"`
	TeamID       uint      `json:"team_id" binding:"required"`
	AssigneeIDs  []uint    `json:"assignee_ids"`
	ParentTaskID *uint     `json:"parent_task_id"`
	LabelIDs     []uint    `json:"label_ids"`
}
