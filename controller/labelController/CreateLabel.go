package labelController

import (
	"net/http"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func CreateLabel(c *gin.Context) {
	var label model.Label
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

	if !isAdmin {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Only admin can create label"})
		return
	}

	if err := c.ShouldBindJSON(&label); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to bind input " + err.Error()})
		return
	}

	if err := DB.Create(&label).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create label " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"label": label})

	// new label notification

	var n model.Notification
	n.Content = "You have Created new " + label.Name + " label."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
}
