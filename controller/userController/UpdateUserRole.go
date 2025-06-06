package userController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func UpdateUserRole(c *gin.Context) {
	DB := config.DB

	val, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": "Unable to get userID from context"})
		return
	}
	requesterID := val.(uint)

	var adminUser model.User
	if err := DB.First(&adminUser, requesterID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch admin user"})
		return
	}
	if adminUser.Role != "Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can change roles"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	newRole, ok := input["role"].(string)
	if !ok || newRole == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role field is missing or invalid"})
		return
	}

	idParam := c.Param("id")
	targetUserID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID in URL"})
		return
	}

	var user model.User
	if err := DB.First(&user, targetUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := DB.Model(&user).Update("role", newRole).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role updated", "user": user})

	// Update Notification
	var notification model.Notification
	notification.Content = adminUser.Name + " has changed your role from " + user.Role + " to " + newRole + "."
	notification.UserID = uint(targetUserID)
	if ok := notificationController.CreateNotification(notification); !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	notification.Content = "You have changed " + user.Name + "'s role from " + user.Role + " to " + newRole + "."
	notification.UserID = adminUser.ID
	if ok := notificationController.CreateNotification(notification); !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
}
