package taskController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func RemoveTaskAssignees(c *gin.Context) {
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the creator or an admin can update the task"})
		return
	}

	var input assignee
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to bind input: " + err.Error()})
		return
	}

	var remainingAssigneeIDs []uint
	var removedAssigneeIDs []uint
	for _, user := range task.Assignees {
		toRemove := false
		for _, removeID := range input.Assignees {
			if user.ID == removeID {
				toRemove = true
				break
			}
		}
		if !toRemove {
			remainingAssigneeIDs = append(remainingAssigneeIDs, user.ID)
		} else {
			removedAssigneeIDs = append(removedAssigneeIDs, user.ID)
		}
	}

	var remainingAssignees []model.User
	if err := DB.Where("id IN ?", remainingAssigneeIDs).Find(&remainingAssignees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch remaining assignees: " + err.Error()})
		return
	}

	var removedAssignees []model.User
	if err := DB.Where("id IN ?", removedAssigneeIDs).Find(&removedAssignees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch remaining assignees: " + err.Error()})
		return
	}

	if err := DB.Model(&task).Association("Assignees").Replace(&remainingAssignees); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update assignees: " + err.Error()})
		return
	}

	DB.Preload("Assignees").Preload("Labels").First(&task, task.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Task assignees updated", "task": task})

	// remove assignees notification

	if len(task.Assignees) > 0 {

		var notifications []model.Notification
		for _, assignee := range removedAssignees {
			var n model.Notification
			n.Content = "You are no longer assignee of " + task.Title + " task."
			n.UserID = assignee.ID
			notifications = append(notifications, n)
		}

		if !notificationController.CreateNotifications(notifications) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
		}
	}
	var n model.Notification
	n.Content = "You have removed assignees from " + task.Title + " task."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
	}
}
