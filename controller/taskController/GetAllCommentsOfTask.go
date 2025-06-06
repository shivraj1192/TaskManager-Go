package taskController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/controller/commentController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func GetAllCommentsOfTask(c *gin.Context) {
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

	var count int64
	if err := DB.Table("task_assignees").
		Where("user_id = ? AND task_id = ?", userID, task.ID).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to check task membership: " + err.Error()})
		return
	}

	if count == 0 && task.CreatorID != userID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only task members, Admin, and creator can view comments of this task"})
		return
	}

	var comments []model.Comment
	if err := DB.Preload("SubComments").Where("task_id = ? AND parent_comment_id = 0", task.ID).Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch comments: " + err.Error()})
		return
	}

	for i := range comments {
		if err := commentController.LoadSubComments(&comments[i]); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load subcomments"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}
