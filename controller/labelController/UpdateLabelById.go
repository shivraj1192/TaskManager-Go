package labelController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func UpdateLabelById(c *gin.Context) {
	DB := config.DB

	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user"})
		return
	}
	userID := userIDRaw.(uint)

	var user model.User
	if err := DB.First(&user, userID).Error; err != nil || user.Role != "Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can view all the labels:" + err.Error()})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid label id:" + err.Error()})
		return
	}

	labelId := uint(id)

	var label model.Label
	if err := DB.Preload("Tasks").First(&label, labelId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Unable to get label:" + err.Error()})
		return
	}

	var input model.Label
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to bind input:" + err.Error()})
		return
	}

	label.Name = input.Name

	if err := DB.Save(&label).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update label:" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"label": label})

	// new label notification

	var n model.Notification
	n.Content = "You have Updated details of " + label.Name + " label."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
}
