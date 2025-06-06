package commentController

import (
	"net/http"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func GetAllComments(c *gin.Context) {
	DB := config.DB

	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user"})
		return
	}
	userID := userIDRaw.(uint)

	var user model.User
	if err := DB.First(&user, userID).Error; err != nil || user.Role != "Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can view all the comments"})
		return
	}

	var comments []model.Comment
	if err := DB.Preload("SubComments").Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch comments: " + err.Error()})
		return
	}

	for i := range comments {
		if err := LoadSubComments(&comments[i]); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load subcomments"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

func LoadSubComments(comment *model.Comment) error {
	if err := config.DB.Preload("SubComments").First(comment, comment.ID).Error; err != nil {
		return err
	}

	for i := range comment.SubComments {
		if err := LoadSubComments(&comment.SubComments[i]); err != nil {
			return err
		}
	}
	return nil
}
