package taskController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func GetTaskById(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to convert string into int " + err.Error()})
		return
	}

	taskId := uint(val)
	var task model.Task
	if err := DB.Preload("Assignees").
		Preload("SubTasks").
		Preload("Labels").
		Preload("Comments").
		Preload("Attachments").
		First(&task, taskId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.CreatorID == userId || isAdmin {
		c.JSON(http.StatusOK, gin.H{"task": task})
	} else {
		c.JSON(http.StatusConflict, gin.H{"error": "only admin or task creator can access team"})
	}
}
