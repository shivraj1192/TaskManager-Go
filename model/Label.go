package model

import "gorm.io/gorm"

type Label struct {
	gorm.Model
	Name  string `json:"name" binding:"required" gorm:"unique"`
	Tasks []Task `gorm:"many2many:task_labels;constraint:OnDelete:CASCADE"`
}
