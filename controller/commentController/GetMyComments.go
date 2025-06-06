package commentController

import (
	"net/http"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func GetMyComments(c *gin.Context) {
	DB := config.DB

	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user"})
		return
	}
	userID := userIDRaw.(uint)

	var comments []model.Comment
	if err := DB.Preload("SubComments").Where("user_id = ?", userID).Find(&comments).Error; err != nil {
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
