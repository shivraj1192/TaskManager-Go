package labelController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func GetLabelById(c *gin.Context) {
	DB := config.DB

	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user"})
		return
	}
	userID := userIDRaw.(uint)

	var user model.User
	if err := DB.First(&user, userID).Error; err != nil || user.Role != "Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can view all the labels"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid label id:" + err.Error()})
		return
	}

	labelId := uint(id)

	var label model.User
	if err := DB.Preload("Tasks").First(&label, labelId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Unable to get label:" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"label": label})
}
