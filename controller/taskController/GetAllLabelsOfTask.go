package taskController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func GetAllLabelsOfTask(c *gin.Context) {
	DB := config.DB

	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user"})
		return
	}
	userID := userIDRaw.(uint)

	isAdmin := false
	var user model.User
	if err := DB.First(&user, userID).Error; err == nil && user.Role == "Admin" {
		isAdmin = true
	}

	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var task model.Task
	if err := DB.First(&task, uint(taskID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var count int64
	if err := DB.Table("task_assignees").
		Where("user_id = ? AND task_id = ?", userID, task.ID).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to check task membership: " + err.Error()})
		return
	}

	if count == 0 && task.CreatorID != userID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only task members, Admin, and creator can view labels of this task"})
		return
	}

	var labels []model.Label
	if err := DB.Joins("JOIN task_labels ON task_labels.label_id = labels.id").
		Where("task_labels.task_id = ?", task.ID).
		Find(&labels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch labels: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"labels": labels})
}
