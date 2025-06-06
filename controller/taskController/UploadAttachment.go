package taskController

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func UploadAttachment(c *gin.Context) {
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

	var count int64
	if err := DB.Table("task_assignees").
		Where("user_id = ? AND task_id = ?", userID, task.ID).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to check task membership: " + err.Error()})
		return
	}

	if count == 0 && task.CreatorID != userID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only task members, Admin, and creator can upload attachments to this task"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required: " + err.Error()})
		return
	}

	saveDir := "./static/file"
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create directory: " + err.Error()})
		return
	}

	uniqueName := file.Filename
	filePath := filepath.Join(saveDir, uniqueName)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save file: " + err.Error()})
		return
	}

	attachment := model.Attachment{
		FileName:   file.Filename,
		URL:        "/static/file/" + uniqueName,
		TaskID:     task.ID,
		UploaderID: user.ID,
	}

	var countAttachment int64
	if err := DB.Model(model.Attachment{}).Where("file_name = ?", attachment.FileName).Count(&countAttachment).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Attachment not found"})
		return
	}

	if countAttachment > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to add same file in one task"})
		return
	}

	if err := DB.Create(&attachment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save attachment in DB"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"attachment": attachment})

	// new attachment notification
	if len(task.Assignees) > 0 {
		var notifications []model.Notification
		for _, assignee := range task.Assignees {
			var n model.Notification
			n.Content = "New attachment uploaded by " + user.Name + " to " + task.Title + " task."
			n.UserID = assignee.ID
			notifications = append(notifications, n)
		}

		if !notificationController.CreateNotifications(notifications) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
			return
		}
	}

	var n model.Notification
	n.Content = "You have uploaded new attachment to " + task.Title + " task."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
}
