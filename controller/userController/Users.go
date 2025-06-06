package userController

import (
	"net/http"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func Users(c *gin.Context) {
	var users []model.User
	DB := config.DB
	var user model.User

	val, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": "Unable to get userId to detect admin or not"})
		return
	}
	userId := val.(uint)

	if err := DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to fetch user to detect admin or not"})
		return
	}
	role := user.Role

	if role == "admin" || role == "Admin" {
		if err := DB.
			Preload("Teams").
			Preload("TasksCreated").
			Preload("TasksAssigned").
			Preload("Comments").
			Preload("Attachments").
			Preload("Notifications").
			Find(&users).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"users": users})

	} else {
		c.JSON(http.StatusConflict, gin.H{"error": "Only admin can see all users list"})
		return
	}

}
