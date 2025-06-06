package taskController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func RemoveTaskLabels(c *gin.Context) {
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
	if err := DB.Preload("Labels").First(&task, uint(taskID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.CreatorID != userID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the creator or an admin can update the task"})
		return
	}

	var input LabelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to bind input: " + err.Error()})
		return
	}

	keepLabelMap := make(map[uint]bool)
	for _, label := range task.Labels {
		keepLabelMap[label.ID] = true
	}
	for _, id := range input.LabelIDs {
		delete(keepLabelMap, id)
	}

	var remainingIDs []uint
	for id := range keepLabelMap {
		remainingIDs = append(remainingIDs, id)
	}

	var remainingLabels []model.Label
	if len(remainingIDs) > 0 {
		if err := DB.Where("id IN ?", remainingIDs).Find(&remainingLabels).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch labels: " + err.Error()})
			return
		}
	}

	if len(remainingIDs) != len(remainingLabels) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input (label id's)"})
		return
	}

	if err := DB.Model(&task).Association("Labels").Replace(remainingLabels); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update labels: " + err.Error()})
		return
	}

	DB.Preload("Labels").Preload("Assignees").First(&task, task.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Task labels updated", "task": task})

	// remove task label notification

	if len(task.Assignees) > 0 {
		var notifications []model.Notification
		for _, assignee := range task.Assignees {
			var n model.Notification
			n.Content = task.Title + " task deleted by " + user.Name + "."
			n.UserID = assignee.ID
			notifications = append(notifications, n)
		}

		if !notificationController.CreateNotifications(notifications) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
			return
		}
	}

	var n model.Notification
	n.Content = "You have deleted " + task.Title + " task."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
}
