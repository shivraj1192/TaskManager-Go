package attachmentcontroller

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func DeleteAttachmentById(c *gin.Context) {
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

	attachID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attachment ID"})
		return
	}

	var attachment model.Attachment
	if err := DB.First(&attachment, attachID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attachment not found"})
		return
	}

	var task model.Task
	if err := DB.First(&task, attachment.TaskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if attachment.UploaderID != userID && task.CreatorID != userID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: only uploader, task creator, or admin can delete this attachment"})
		return
	}

	if err := DB.Unscoped().Delete(&attachment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete attachment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "Attachment got deleted successfully"})

	// Delete attachment

	var n model.Notification
	n.Content = "You have deleted " + attachment.FileName + " attachment."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
}
