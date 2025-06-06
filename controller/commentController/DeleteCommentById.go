package commentController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func DeleteCommentById(c *gin.Context) {
	DB := config.DB

	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user"})
		return
	}
	userID := userIDRaw.(uint)

	var user model.User
	if err := DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to Identify user"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task id:" + err.Error()})
		return
	}

	commentId := uint(id)

	var comment model.Comment
	if err := DB.First(&comment, commentId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch comments: " + err.Error()})
		return
	}

	var task model.Task
	if err := DB.First(&task, comment.TaskID).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unable to recognize comment's task:" + err.Error()})
		return
	}

	if comment.UserID != userID && task.CreatorID != userID && user.Role != "Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You dont have authorization to delee the comment"})
		return
	}

	var count int64
	if err := DB.Model(&model.Comment{}).Where("parent_comment_id = ?", commentId).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Unable to get comments of parent id"})
		return
	}

	if count != 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Can't delete parent comments"})
		return
	}

	if err := DB.Unscoped().Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"seccess": "Comment Deleted successfully"})

}
