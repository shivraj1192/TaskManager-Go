package taskController

import (
	"net/http"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func GetMyCreatedTasks(c *gin.Context) {
	userId, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	userID := userId.(uint)

	DB := config.DB

	var tasks []model.Task
	if err := DB.Where("creator_id = ?", userID).
		Preload("Assignees").
		Preload("SubTasks").
		Preload("Labels").
		Preload("Comments").
		Preload("Attachments").
		Find(&tasks).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}
