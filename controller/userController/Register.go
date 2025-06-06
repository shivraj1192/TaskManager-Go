package userController

import (
	"net/http"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var user model.User
	var existUser model.User
	var hasedPassword []byte
	var err error
	db := config.DB
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to bind with user" + err.Error()})
		return
	}
	if err := db.Where("email=?", user.Email).First(&existUser); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	if hasedPassword, err = bcrypt.GenerateFromPassword([]byte(user.Password), 14); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to hash password"})
		return
	}
	user.Password = string(hasedPassword)

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to create user" + user.Name + "\n" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "Created user" + user.Name, "user": user})

	// User register Notification
	var notification model.Notification
	notification.Content = "Welcome to Application."
	notification.UserID = user.ID
	if ok := notificationController.CreateNotification(notification); !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
}
