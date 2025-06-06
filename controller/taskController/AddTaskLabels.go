package taskController

import (
	"fmt"
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

type LabelInput struct {
	LabelIDs []uint `json:"labels"`
}

func AddTaskLabels(c *gin.Context) {
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
	if err := DB.Preload("Labels").First(&task, uint(taskID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.CreatorID != userID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the creator or an admin can update the task"})
		return
	}

	var input LabelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to bind input: " + err.Error()})
		return
	}

	labelIDSet := make(map[uint]struct{})
	for _, label := range task.Labels {
		labelIDSet[label.ID] = struct{}{}
	}
	for _, id := range input.LabelIDs {
		labelIDSet[id] = struct{}{}
	}

	var mergedLabelIDs []uint
	for id := range labelIDSet {
		mergedLabelIDs = append(mergedLabelIDs, id)
	}

	var mergedLabels []model.Label
	if err := DB.Where("id IN ?", mergedLabelIDs).Find(&mergedLabels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch labels: " + err.Error()})
		return
	}

	if len(mergedLabelIDs) != len(mergedLabels) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid input, enter correct label id's"})
		return
	}

	if err := DB.Model(&task).Association("Labels").Replace(mergedLabels); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update labels: " + err.Error()})
		return
	}

	DB.Preload("Labels").Preload("Assignees").First(&task, task.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Task labels updated", "task": task})

	// New labels notification

	if len(task.Assignees) > 0 {
		var notifications []model.Notification
		for _, assignee := range task.Assignees {
			var n model.Notification
			n.Content = user.Name + " added new labels to " + task.Title + " task."
			n.UserID = assignee.ID
			notifications = append(notifications, n)
		}

		fmt.Println(notifications)

		if !notificationController.CreateNotifications(notifications) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
		}
	}
	var n model.Notification
	n.Content = "You have added labels of " + task.Title + " task."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
	}
}
