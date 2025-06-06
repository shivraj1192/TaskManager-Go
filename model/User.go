package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Uname         string         `gorm:"unique" json:"uname" binding:"required"`
	Name          string         `json:"name" binding:"required"`
	Email         string         `gorm:"unique" json:"email" binding:"required"`
	Password      string         `json:"password" binding:"required"`
	Role          string         `gorm:"default:'Member'" json:"role"`
	Teams         []Team         `gorm:"many2many:team_members;constraint:OnDelete:CASCADE" json:"teams"`
	TasksCreated  []Task         `gorm:"foreignKey:CreatorID;constraint:OnDelete:CASCADE" json:"tasks_created"`
	TasksAssigned []Task         `gorm:"many2many:task_assignees;constraint:OnDelete:CASCADE" json:"tasks_assigned"`
	Comments      []Comment      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"comments"`
	Attachments   []Attachment   `gorm:"foreignKey:UploaderID;constraint:OnDelete:CASCADE" json:"attachments"`
	Notifications []Notification `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"notifications"`
}

type LoginUser struct {
	Uname    string `jaon:"uname"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required"`
}
