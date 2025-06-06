package model

import "gorm.io/gorm"

type Comment struct {
	gorm.Model
	Content         string    `json:"content" binding:"required"`
	UserID          uint      `json:"user_id"`
	TaskID          uint      `json:"task_id"`
	ParentCommentID *uint     `json:"parent_comment_id"`
	SubComments     []Comment `gorm:"foreignKey:ParentCommentID;constraint:OnDelete:CASCADE" json:"subcomments"`
}
