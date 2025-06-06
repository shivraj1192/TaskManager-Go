package taskController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

type assignee struct {
	Assignees []uint `json:"assignees"`
}

func AddTaskAssignees(c *gin.Context) {
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

	var input assignee
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to bind input: " + err.Error()})
		return
	}

	var validAssigneeIDs []uint
	for _, uid := range input.Assignees {
		var count int64
		DB.Table("team_members").
			Where("team_id = ? AND user_id = ?", task.TeamID, uid).
			Count(&count)
		if count == 1 {
			validAssigneeIDs = append(validAssigneeIDs, uid)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "All users should be task's team member"})
			return
		}
	}

	var existingAssigneeIDs []uint
	for _, user := range task.Assignees {
		existingAssigneeIDs = append(existingAssigneeIDs, user.ID)
	}

	allAssigneeMap := make(map[uint]struct{})
	for _, id := range append(existingAssigneeIDs, validAssigneeIDs...) {
		allAssigneeMap[id] = struct{}{}
	}

	var uniqueIDs []uint
	for id := range allAssigneeMap {
		uniqueIDs = append(uniqueIDs, id)
	}

	var newAssignees []model.User
	if err := DB.Where("id IN ?", uniqueIDs).Find(&newAssignees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignees: " + err.Error()})
		return
	}

	if err := DB.Model(&task).Association("Assignees").Replace(&newAssignees); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update assignees: " + err.Error()})
		return
	}

	DB.Preload("Assignees").First(&task, task.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Task assignees updated", "task": task})

	// Add Assignee Notification

	var notifications []model.Notification
	for _, u := range newAssignees {
		var n model.Notification
		n.Content = "You have beed assigned to " + task.Title + " task by " + user.Name + "."
		n.UserID = u.ID
		notifications = append(notifications, n)
	}
	if !notificationController.CreateNotifications(notifications) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
		return
	}

	var n model.Notification
	n.Content = "You have assigned members to " + task.Title + " task."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
}
