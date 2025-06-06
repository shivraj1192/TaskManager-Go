package teamController

import (
	"net/http"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func GetAllTeams(c *gin.Context) {
	var teams []model.Team
	DB := config.DB

	id, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to recognize you"})
		return
	}
	userId := id.(uint)
	role := "Admin"

	var user model.User
	if err := DB.Where("ID = ? AND role=?", userId, role).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only admin view all teams"})
		return
	}

	if err := DB.Preload("Members").Preload("Tasks").Find(&teams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch teams" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}
