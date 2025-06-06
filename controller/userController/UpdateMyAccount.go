package userController

import (
	"net/http"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func UpdateMyAccount(c *gin.Context) {
	var user model.User
	DB := config.DB
	val, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": "Unable get userId from token"})
		return
	}
	userId := val.(uint)

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "unable to bind input"})
		return
	}

	if _, ok := input["password"]; ok {
		c.JSON(http.StatusConflict, gin.H{"error": "Cant update password here try Updatepassword api"})
		return
	}

	if _, ok := input["role"]; ok {
		c.JSON(http.StatusConflict, gin.H{"error": "only admin can change role"})
		return
	}

	if err := DB.Where("ID=?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to fetch user" + err.Error()})
		return
	}

	if err := DB.Model(&user).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update user" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})

	// Update Notification
	var notification model.Notification
	notification.Content = "You have updated your details. If not change your password."
	notification.UserID = userId
	if ok := notificationController.CreateNotification(notification); !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

}
