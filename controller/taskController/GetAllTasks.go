package taskController

import (
	"net/http"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func GetAllTasks(c *gin.Context) {
	var tasks []model.Task
	DB := config.DB

	id, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to recognize you"})
		return
	}
	userId := id.(uint)
	role := "Admin"

	var user model.User
	if err := DB.Where("ID = ? AND role=?", userId, role).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only admin view all teams"})
		return
	}

	if err := DB.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tasks not found"})
		return
	}

	for i := range tasks {
		if err := loadSubTasks(&tasks[i]); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load subtasks"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func loadSubTasks(task *model.Task) error {
	if err := config.DB.Preload("Assignees").
		Preload("Labels").
		Preload("Comments").
		Preload("Attachments").
		Preload("SubTasks").
		First(task, task.ID).Error; err != nil {
		return err
	}

	for i := range task.SubTasks {
		if err := loadSubTasks(&task.SubTasks[i]); err != nil {
			return err
		}
	}
	return nil
}
