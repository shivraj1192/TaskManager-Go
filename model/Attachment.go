package model

import "gorm.io/gorm"

type Attachment struct {
	gorm.Model
	FileName   string `json:"file_name" binding:"required"`
	URL        string `json:"url" binding:"required"`
	TaskID     uint   `json:"task_id"`
	UploaderID uint   `json:"uploader_id"`
}
