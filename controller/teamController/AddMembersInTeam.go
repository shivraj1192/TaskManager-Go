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

func AddMembersInTeam(c *gin.Context) {
	var members model.Members
	var users []model.User
	var team model.Team
	DB := config.DB

	// fmt.Println("Start")
	id, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to recognize you"})
		return
	}
	userId := id.(uint)

	// fmt.Println("Start2")

	isAdmin := false
	var user model.User
	if err := DB.Where("ID = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User Not Found: " + err.Error()})
		return
	}
	if user.Role == "Admin" {
		isAdmin = true
	}

	// fmt.Println("Start3")

	val, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to fetch team id from request: " + err.Error()})
		return
	}
	teamId := uint(val)

	// fmt.Println("Start4")

	if err := c.ShouldBindJSON(&members); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if err := DB.Where("ID IN ?", members.UserIds).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch users: " + err.Error()})
		return
	}

	if err := DB.Preload("Members").First(&team, teamId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get team: " + err.Error()})
		return
	}

	if team.OwnerId != userId && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team owner or admin can add members"})
		return
	}

	existingMemberIDs := map[uint]bool{}
	for _, member := range team.Members {
		existingMemberIDs[member.ID] = true
	}

	var newUsers []model.User
	for _, u := range users {
		if !existingMemberIDs[u.ID] {
			newUsers = append(newUsers, u)
		}
	}

	if len(newUsers) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No new members to add", "team": team})
		return
	}

	if err := DB.Model(&team).Association("Members").Append(&newUsers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to add new members: " + err.Error()})
		return
	}

	if err := DB.Preload("Members").First(&team, teamId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch updated team: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "New members added", "team": team})

	// Add Notifications
	memberString := ""
	var notifications []model.Notification
	for _, u := range newUsers {
		var n model.Notification
		n.Content = "You have beed added to " + team.Name + " team by " + user.Name + "."
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
		n.Content = "New members (" + memberString + ") got added to " + team.Name + " by " + user.Name + "."
		n.UserID = member.ID
		notifications = append(notifications, n)
	}
	if !notificationController.CreateNotifications(notifications) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
		return
	}

	var n model.Notification
	n.Content = "You have added (" + memberString + ") members to " + team.Name + " team."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
}
