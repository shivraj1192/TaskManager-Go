package commentController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func GetCommentById(c *gin.Context) {
	DB := config.DB

	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user"})
		return
	}
	userID := userIDRaw.(uint)

	var user model.User
	if err := DB.First(&user, userID).Error; err != nil || user.Role != "Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can view comments by id"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task id:" + err.Error()})
		return
	}

	commentId := uint(id)

	var comment model.Comment
	if err := DB.Preload("SubComments").First(&comment, commentId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch comments: " + err.Error()})
		return
	}

	if err := LoadSubComments(&comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load subcomments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comment": comment})
}
