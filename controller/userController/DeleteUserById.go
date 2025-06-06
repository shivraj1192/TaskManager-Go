package userController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func DeleteUserById(c *gin.Context) {
	val, err := strconv.Atoi(c.Param("id"))
	var user model.User
	DB := config.DB
	id, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": "Unable to get userID from context"})
		return
	}
	requesterID := id.(uint)
	var adminUser model.User
	if err := DB.First(&adminUser, requesterID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch admin user"})
		return
	}
	if adminUser.Role != "Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can delete users"})
		return
	}

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"Error": "cant convert string id into int"})
		return
	}

	userId := uint(val)

	if err := DB.Preload("Teams").Where("ID=?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Invalid user_id" + err.Error()})
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

	c.JSON(http.StatusOK, gin.H{"notification": "Notifications deleted successfully"})

	if err := DB.Unscoped().Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Unable to delete" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "Deleted Successfully"})

}
