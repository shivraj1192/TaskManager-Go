package teamController

import (
	"fmt"
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func UpdateTeam(c *gin.Context) {
	var team model.Team
	var owner model.User
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

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to get teamid from api" + err.Error()})
		return
	}
	teamId := uint(teamID)

	if err := DB.Preload("Members").Where("ID=?", teamId).First(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to get team" + err.Error()})
		return
	}

	if team.OwnerId == userId || isAdmin {
		var input map[string]interface{}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Team input Binding error"})
			return
		}

		if val, ok := input["owner_id"]; ok {
			ownerId := uint(val.(float64))
			ownerPresent := false
			for _, member := range team.Members {
				if member.ID == ownerId {
					ownerPresent = true
				}
			}
			if !ownerPresent {
				fmt.Println(ownerId)

				if err := DB.Where("ID=?", ownerId).First(&owner).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"Error": "Unable to get owner id" + err.Error()})
					return
				}

				if err := DB.Model(&team).Association("Members").Append(&owner); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add owner to team members: " + err.Error()})
					return
				}
			}
		}

		if err := DB.Model(&team).Updates(input).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update team" + err.Error()})
			return
		}
		if err := DB.Preload("Members").First(&team, team.ID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load team members: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"team": team})

		// update notification
		var notifications []model.Notification
		for _, member := range team.Members {
			var n model.Notification
			n.Content = team.Name + " team updated by " + user.Name + "."
			n.UserID = member.ID
			notifications = append(notifications, n)
		}

		if !notificationController.CreateNotifications(notifications) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
			return
		}

		var n model.Notification
		n.Content = "You have updated " + team.Name + "team."
		n.UserID = user.ID
		if !notificationController.CreateNotification(n) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
			return
		}
	} else {
		c.JSON(http.StatusConflict, gin.H{"error": "only team creator or admin can edit details of this team"})
	}

}
