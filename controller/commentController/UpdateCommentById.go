package commentController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func UpdateCommentById(c *gin.Context) {
	DB := config.DB

	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user"})
		return
	}
	userID := userIDRaw.(uint)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment id:" + err.Error()})
		return
	}

	commentId := uint(id)

	var comment model.Comment
	if err := DB.First(&comment, commentId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch comments: " + err.Error()})
		return
	}

	if userID != comment.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized person to update comment"})
		return
	}

	var input model.Comment
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input : " + err.Error()})
		return
	}

	if comment.TaskID != input.TaskID || comment.UserID != input.UserID || comment.ParentCommentID != input.ParentCommentID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can't change userId, ParentCommentId & taskId"})
		return
	}

	comment.Content = input.Content

	if err := DB.Save(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update comment:" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comment": comment})

}
