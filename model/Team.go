package model

import "gorm.io/gorm"

type Team struct {
	gorm.Model
	Name        string `gorm:"unique" json:"name" binding:"required"`
	Description string `json:"description"`
	OwnerId     uint   `json:"owner_id"`
	Members     []User `gorm:"many2many:team_members;constraint:OnDelete:CASCADE" json:"members"`
	Tasks       []Task `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE" json:"tasks"`
}

type Members struct {
	UserIds []uint `json:"userids" binding:"required"`
}
