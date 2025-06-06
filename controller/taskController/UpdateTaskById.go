package taskController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func UpdateTaskById(c *gin.Context) {
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
	if err := DB.Preload("Assignees").
		Preload("Labels").
		Preload("Comments").
		Preload("Attachments").
		Preload("SubTasks").
		First(&task, uint(taskID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.CreatorID != userID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the creator or an admin can update the task"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if cidRaw, ok := input["creator_id"]; ok {
		cidFloat, ok := cidRaw.(float64)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "creator_id must be a number"})
			return
		}
		cid := uint(cidFloat)

		var count int64
		if err := DB.Table("team_members").
			Where("user_id = ? AND team_id = ?", cid, task.TeamID).
			Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate team membership"})
			return
		}
		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "creator is not a team member of the team"})
			return
		}
	}

	if err := DB.Model(&task).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task: " + err.Error()})
		return
	}

	DB.Preload("Assignees").
		Preload("Labels").
		Preload("Comments").
		Preload("Attachments").
		Preload("SubTasks").
		First(&task, uint(taskID))

	c.JSON(http.StatusOK, gin.H{"task": task})

	// update task notifications

	if len(task.Assignees) > 0 {
		var notifications []model.Notification
		for _, assignee := range task.Assignees {
			var n model.Notification
			n.Content = task.Title + " task details are updated by " + user.Name + "."
			n.UserID = assignee.ID
			notifications = append(notifications, n)
		}

		if !notificationController.CreateNotifications(notifications) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
			return
		}
	}
	var n model.Notification
	n.Content = "You have updated details of " + task.Title + " task."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

}
