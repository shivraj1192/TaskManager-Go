package userController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func GetUserById(c *gin.Context) {
	DB := config.DB

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
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin or account holder can view users"})
		return
	}

	var user model.User
	if err := DB.
		Preload("Teams").
		Preload("TasksCreated").
		Preload("TasksAssigned").
		Preload("Comments").
		Preload("Attachments").
		Preload("Notifications").
		Find(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}
