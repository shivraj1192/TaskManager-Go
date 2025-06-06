package taskController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func UpdateTaskTeamById(c *gin.Context) {
	DB := config.DB

	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user"})
		return
	}
	userID := userIDRaw.(uint)

	isAdmin := false
	var user model.User
	if err := DB.First(&user, userID).Error; err == nil && user.Role == "Admin" {
		isAdmin = true
	}

	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var task model.Task
	if err := DB.Preload("Assignees").First(&task, uint(taskID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.CreatorID != userID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the creator or an admin can update the task"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	teamIDRaw, ok := input["team_id"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing team_id"})
		return
	}
	teamIDFloat, ok := teamIDRaw.(float64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team_id must be a number"})
		return
	}
	newTeamID := uint(teamIDFloat)

	var count int64
	if err := DB.Table("team_members").
		Where("team_id = ? AND user_id = ?", newTeamID, task.CreatorID).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate team membership"})
		return
	}
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Creator is not a member of the new team"})
		return
	}

	if err := DB.Model(&task).Update("team_id", newTeamID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team_id"})
		return
	}

	var newTeamMembers []model.User
	if err := DB.Joins("JOIN team_members ON team_members.user_id = users.id").
		Where("team_members.team_id = ?", newTeamID).
		Find(&newTeamMembers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch new team members"})
		return
	}

	teamMemberMap := make(map[uint]bool)
	for _, member := range newTeamMembers {
		teamMemberMap[member.ID] = true
	}

	var updatedAssignees []model.User
	for _, assignee := range task.Assignees {
		if teamMemberMap[assignee.ID] {
			updatedAssignees = append(updatedAssignees, assignee)
		}
	}

	if err := DB.Model(&task).Association("Assignees").Replace(&updatedAssignees); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update assignees"})
		return
	}

	DB.Preload("Assignees").First(&task, task.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Task team and assignees updated", "task": task})

	// update task's teamId notification
	if len(task.Assignees) > 0 {
		var notifications []model.Notification
		for _, assignee := range task.Assignees {
			var n model.Notification
			n.Content = task.Title + " task's team is changed by " + user.Name + "."
			n.UserID = assignee.ID
			notifications = append(notifications, n)
		}

		if !notificationController.CreateNotifications(notifications) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
			return
		}
	}
	var n model.Notification
	n.Content = "You have changed team of " + task.Title + " task."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
}
