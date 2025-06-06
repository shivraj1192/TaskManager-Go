package labelController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func DeleteLabelById(c *gin.Context) {
	DB := config.DB

	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user"})
		return
	}
	userID := userIDRaw.(uint)

	var user model.User
	if err := DB.First(&user, userID).Error; err != nil || user.Role != "Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can delete labels"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid label id:" + err.Error()})
		return
	}

	labelId := uint(id)

	var label model.User
	if err := DB.Preload("Tasks").First(&label, labelId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Unable to get label:" + err.Error()})
		return
	}

	if err := DB.Model(&label).Association("Tasks").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cant clear the asscociation of label:" + err.Error()})
		return
	}

	if err := DB.Unscoped().Delete(&label).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete label:" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "Labe got deleted successfully"})

	// delete label notification

	var n model.Notification
	n.Content = "You have deleted " + label.Name + " label."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
}
