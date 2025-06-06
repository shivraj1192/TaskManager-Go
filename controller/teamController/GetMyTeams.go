package teamController

import (
	"net/http"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func GetMyTeams(c *gin.Context) {
	userId, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	userID := userId.(uint)

	DB := config.DB

	var teams []model.Team
	if err := DB.
		Joins("JOIN team_members ON team_members.team_id = teams.id").
		Where("team_members.user_id = ?", userID).
		Preload("Members").
		Preload("Tasks").
		Find(&teams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch user teams: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}
