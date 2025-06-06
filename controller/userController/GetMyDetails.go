package userController

import (
	"net/http"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func GetMyDetails(c *gin.Context) {
	var user model.User
	DB := config.DB

	val, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't fetch user from token"})
		return
	}

	userId := val.(uint)
	if err := DB.
		Preload("Teams").
		Preload("TasksCreated").
		Preload("TasksAssigned").
		Preload("Comments").
		Preload("Attachments").
		Preload("Notifications").
		Find(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}
