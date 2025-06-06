package teamController

import (
	"net/http"
	"strconv"
	"strings"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func RemoveMembersFromTeam(c *gin.Context) {
	var members model.Members
	var users []model.User
	var team model.Team
	DB := config.DB

	isAdmin := false
	id, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to recognize you"})
		return
	}
	userId := id.(uint)
	role := "Admin"

	var user model.User
	if err := DB.Where("ID = ? AND role=?", userId, role).First(&user).Error; err == nil {
		isAdmin = true
	}

	val, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to fetch team id from request" + err.Error()})
		return
	}
	teamId := uint(val)

	if err := c.ShouldBindJSON(&members); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to fetch members from request" + err.Error()})
		return
	}

	if err := DB.Where("ID IN ?", members.UserIds).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch users from members" + err.Error()})
		return
	}

	if err := DB.Preload("Members").First(&team, teamId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get team after append" + err.Error()})
		return
	}

	if team.OwnerId == userId || isAdmin {
		if len(members.UserIds) != len(users) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid member entries"})
			return
		}

		for _, member := range users {
			if team.OwnerId == member.ID {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't remove owner of team"})
				return
			}
		}

		if err := DB.Model(&team).Association("Members").Delete(&users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to remove users" + err.Error()})
			return
		}

		if err := DB.Preload("Members").First(&team, teamId).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get team after remove" + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"team": team})

		// Removed members notification

		memberString := ""
		var notifications []model.Notification
		for _, u := range users {
			var n model.Notification
			n.Content = "You have beed removed from " + team.Name + " team by " + user.Name + "."
			n.UserID = u.ID
			notifications = append(notifications, n)
			memberString += u.Name + ", "
		}
		memberString = strings.TrimSuffix(memberString, ", ")
		if !notificationController.CreateNotifications(notifications) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
			return
		}

		notifications = nil
		for _, member := range team.Members {
			var n model.Notification
			n.Content = "(" + memberString + ") this members got removed from " + team.Name + " by " + user.Name + "."
			n.UserID = member.ID
			notifications = append(notifications, n)
		}
		if !notificationController.CreateNotifications(notifications) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
			return
		}

		var n model.Notification
		n.Content = "You have removed (" + memberString + ") members from " + team.Name + " team."
		n.UserID = user.ID
		if !notificationController.CreateNotification(n) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
			return
		}

	} else {
		c.JSON(http.StatusOK, gin.H{"error": "only team creator or admin can remove members"})
	}

}
