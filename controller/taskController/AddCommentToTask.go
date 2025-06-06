package taskController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/commentController"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func AddCommentToTask(c *gin.Context) {
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
	if err := DB.Preload("Comments").First(&task, uint(taskID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var count int64
	if err := DB.Table("task_assignees").
		Where("user_id = ? AND task_id = ?", userID, task.ID).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unable to find user: " + err.Error()})
		return
	}

	if count == 0 && task.CreatorID != userID && !isAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only task members, Admin and creator can comment on this task"})
		return
	}

	var comment model.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to bind input comment: " + err.Error()})
		return
	}

	comment.UserID = userID
	comment.TaskID = task.ID

	if comment.ParentCommentID != nil && *comment.ParentCommentID == 0 {
		comment.ParentCommentID = nil
	}

	if err := DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create comment: " + err.Error()})
		return
	}

	if err := config.DB.Preload("Assignees").
		Preload("Labels").
		Preload("Comments").
		Preload("Attachments").
		Preload("SubTasks").
		First(&task, task.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get task: " + err.Error()})
		return
	}

	for i := range task.Comments {
		if err := commentController.LoadSubComments(&task.Comments[i]); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load subcomments"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"comment": comment, "task": task})

	// New Comment Notification

	var notifications []model.Notification
	for _, assignee := range task.Assignees {
		var n model.Notification
		n.Content = user.Name + " added new comment to " + task.Title + " task."
		n.UserID = assignee.ID
		notifications = append(notifications, n)
	}

	if !notificationController.CreateNotifications(notifications) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
		return
	}

	var n model.Notification
	n.Content = "You have commented on " + task.Title + " task."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
}
