package taskController

import (
	"net/http"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func CreateTask(c *gin.Context) {
	var input model.CreateTaskInput
	DB := config.DB

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to bind input " + err.Error()})
		return
	}

	userID := c.GetUint("userID")

	isAdmin := false
	var user model.User
	if err := DB.First(&user, userID).Error; err == nil && user.Role == "Admin" {
		isAdmin = true
	}

	var team model.Team
	if err := DB.First(&team, input.TeamID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Team is invalid " + err.Error()})
		return
	}

	var count int64
	DB.Table("team_members").
		Where("team_id = ? AND user_id = ?", team.ID, user.ID).
		Count(&count)
	if count == 0 && !isAdmin {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Only team member or admin can create task"})
		return
	}

	unique := map[uint]struct{}{}
	for _, id := range input.AssigneeIDs {
		unique[id] = struct{}{}
	}
	unique[userID] = struct{}{}

	ids := make([]uint, 0, len(unique))
	for id := range unique {
		ids = append(ids, id)
	}

	var assignees []model.User
	if len(ids) > 0 {
		if err := DB.Where("id IN ?", ids).Find(&assignees).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get assignees " + err.Error()})
			return
		}
	}

	count = 0
	if err := DB.Table("team_members").
		Where("team_id = ? AND user_id IN ?", input.TeamID, ids).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}
	if count != int64(len(ids)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All assignees and task creator must be part of the team"})
		return
	}
	var labels []model.Label
	if len(input.LabelIDs) > 0 {
		if err := DB.Where("id IN ?", input.LabelIDs).Find(&labels).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get labels " + err.Error()})
			return
		}
	}

	var dupCheck model.Task
	if err := DB.Where("creator_id = ? AND title = ?", userID, input.Title).First(&dupCheck).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You can't create task of same title with same creator"})
		return
	}

	if input.ParentTaskID != nil && *input.ParentTaskID == 0 {
		input.ParentTaskID = nil
	}

	if input.ParentTaskID != nil {
		var parent model.Task
		if err := DB.First(&parent, *input.ParentTaskID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parent task not found"})
			return
		}
		if parent.TeamID != input.TeamID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parent task must belong to the same team"})
			return
		}
	}

	task := model.Task{
		Title:        input.Title,
		Description:  input.Description,
		Status:       input.Status,
		Priority:     input.Priority,
		DueDate:      input.DueDate,
		CreatorID:    userID,
		TeamID:       input.TeamID,
		Assignees:    assignees,
		ParentTaskID: input.ParentTaskID,
		Labels:       labels,
	}
	if err := DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create task " + err.Error()})
		return
	}

	if err := DB.Preload("Assignees").
		Preload("SubTasks").
		Preload("Labels").
		Preload("Comments").
		Preload("Attachments").
		First(&task, task.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})

	// New task notification

	var notifications []model.Notification
	for _, assignee := range task.Assignees {
		var n model.Notification
		n.Content = "You have been assigned to " + task.Title + " task by " + user.Name + "."
		n.UserID = assignee.ID
		notifications = append(notifications, n)
	}

	if !notificationController.CreateNotifications(notifications) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notifications"})
		return
	}

	var n model.Notification
	n.Content = "You have created new task (" + task.Title + " task)."
	n.UserID = user.ID
	if !notificationController.CreateNotification(n) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
}
