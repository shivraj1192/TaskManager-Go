package userController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func UpdateUserById(c *gin.Context) {
	DB := config.DB
	var user model.User

	idRaw := c.Param("id")
	idInt, err := strconv.Atoi(idRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get id from parameter"})
		return
	}

	id := uint(idInt)

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
	if adminUser.Role != "Admin" && id != adminUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin or account holder can change details"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "unable to bind input"})
		return
	}

	var hashedPassword string
	if password, ok := input["password"]; ok {
		hashedPassword = hashPassword(c, password.(string))
	}
	input["password"] = hashedPassword

	if err := DB.Where("ID=?", id).First(&user).Error; err != nil {
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
	if id == requesterID {
		notification.Content = "You have updated your details. If not change your password."
		notification.UserID = id
		if ok := notificationController.CreateNotification(notification); !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
			return
		}
	} else {
		notification.Content = "Your account is updated by admin " + adminUser.Name
		notification.UserID = id
		if ok := notificationController.CreateNotification(notification); !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
			return
		}

		notification.Content = "You have updated account of user of id " + idRaw
		notification.UserID = adminUser.ID
		if ok := notificationController.CreateNotification(notification); !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
			return
		}
	}

}

func hashPassword(c *gin.Context, password string) string {
	if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return ""
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return ""
	}
	return string(hashedPassword)
}
