package taskController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

type parentTaskIdStr struct {
	ParentTaskID *uint `json:"parent_task_id"`
}

func UpdateParentTaskId(c *gin.Context) {
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
	if err := DB.First(&task, uint(taskID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.CreatorID != userID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the creator or an admin can update the task"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to bind input: " + err.Error()})
		return
	}

	rawParentID, ok := input["parent_task_id"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parent_task_id is required"})
		return
	}

	parentIDFloat, ok := rawParentID.(float64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parent_task_id must be a number"})
		return
	}

	parentTaskID := uint(parentIDFloat)

	if parentTaskID == task.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A task cannot be its own parent"})
		return
	}

	var parentTask model.Task
	if err := DB.First(&parentTask, parentTaskID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parent task not found"})
		return
	}

	if parentTask.TeamID != task.TeamID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parent task must belong to the same team"})
		return
	}

	var parenttaskidstr parentTaskIdStr
	if err := c.ShouldBindJSON(&parenttaskidstr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to bind input: " + err.Error()})
		return
	}

	if *parenttaskidstr.ParentTaskID == 0 {
		parenttaskidstr.ParentTaskID = nil
	}

	task.ParentTaskID = parenttaskidstr.ParentTaskID
	if err := DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update parent task ID: " + err.Error()})
		return
	}

	DB.Preload("Assignees").
		Preload("Labels").
		Preload("Comments").
		Preload("Attachments").
		Preload("SubTasks").
		First(&task, uint(taskID))

	c.JSON(http.StatusOK, gin.H{"task": task})

	// update parent task notification

	if len(task.Assignees) > 0 {
		var notifications []model.Notification
		for _, assignee := range task.Assignees {
			var n model.Notification
			n.Content = task.Title + " task's parent is changed by " + user.Name + "."
			n.UserID = assignee.ID
			notifications = append(notifications, n)
		}

		if !notificationController.CreateNotifications(notifications) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
			return
		}
	}

	var n model.Notification
	n.Content = "You have changed parent of " + task.Title + " task."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
}
