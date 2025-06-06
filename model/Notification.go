package model

import (
	"gorm.io/gorm"
)

type Notification struct {
	gorm.Model
	UserID  uint   `json:"user_id" binding:"required"`
	Content string `json:"content" binding:"required"`
	IsRead  bool   `gorm:"default:false" json:"is_read"`
}
