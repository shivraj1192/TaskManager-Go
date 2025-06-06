package userController

import (
	"net/http"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func DeleteMyAccount(c *gin.Context) {
	id, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": "Unable to get userid from token"})
		return
	}
	userId := id.(uint)

	DB := config.DB
	var user model.User

	if err := DB.Preload("Teams").First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := DB.Model(&user).Association("Teams").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user from teams: " + err.Error()})
		return
	}

	var teams []model.Team
	if err := DB.Where("owner_id=?", userId).Find(&teams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete"})
		return
	}

	for _, team := range teams {
		team.OwnerId = 0
		if err := DB.Save(&team); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to erase your ownership"})
			return
		}
	}

	// Delete notifications
	var notifications []model.Notification
	if err := DB.Where("user_id = ?", userId).Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get user's notifications: " + err.Error()})
		return
	}

	if len(notifications) > 0 {
		if err := DB.Unscoped().Delete(&notifications).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to permanently delete notifications: " + err.Error()})
			return
		}
	}

	if err := DB.Unscoped().Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "User and his/her notifications also deleted successfully"})
}
