package taskController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func DeleteTaskById(c *gin.Context) {
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
	if err := DB.Preload("Assignees").First(&task, uint(taskID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.CreatorID != userID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the creator or an admin can delete the task"})
		return
	}

	var count int64
	if err := DB.Model(&model.Task{}).Where("parent_task_id = ?", task.ID).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to check for subtasks: " + err.Error()})
		return
	}

	if count != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete a task which is a parent of another task"})
		return
	}

	if err := DB.Unscoped().Delete(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "Deleted successfully"})

	// Delete task notification
	var notifications []model.Notification
	for _, assignee := range task.Assignees {
		var n model.Notification
		n.Content = user.Name + " have deleted the " + task.Title + " task."
		n.UserID = assignee.ID
		notifications = append(notifications, n)
	}

	if !notificationController.CreateNotifications(notifications) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
		return
	}

	var n model.Notification
	n.Content = "You have deleted the " + task.Title + " task."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

}
