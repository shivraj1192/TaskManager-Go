package teamController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func DeleteTeamById(c *gin.Context) {
	var team model.Team
	DB := config.DB

	id, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to recognize you"})
		return
	}
	userId := id.(uint)

	var user model.User
	if err := DB.Where("ID = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User Not Found: " + err.Error()})
		return
	}

	val, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to convert string into int"})
		return
	}
	teamId := uint(val)

	if err := DB.Preload("Members").Where("ID=?", teamId).First(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to fetch team: " + err.Error()})
		return
	}

	if userId != team.OwnerId && user.Role != "Admin" {
		c.JSON(http.StatusConflict, gin.H{"error": "Only team creator or admin can delete team"})
		return
	}

	if err := DB.Model(&team).Association("Members").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove team members: " + err.Error()})
		return
	}

	if err := DB.Unscoped().Delete(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete team: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "Team deleted successfully"})

	// Send notifications to members
	var notifications []model.Notification
	for _, assignee := range team.Members {
		var n model.Notification
		n.Content = user.Name + " have deleted the " + team.Name + " team."
		n.UserID = assignee.ID
		notifications = append(notifications, n)
	}

	if !notificationController.CreateNotifications(notifications) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
		return
	}

	var n model.Notification
	n.Content = "You have deleted the " + team.Name + " team."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
}
