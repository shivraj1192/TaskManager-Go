package teamController

import (
	"net/http"
	"strconv"
	"task-manager/config"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func GetTeamById(c *gin.Context) {
	var team model.Team
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

	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "not able to convert string into integer"})
		return
	}
	teamId := uint(teamID)
	if err := DB.Preload("Members").Where("ID=?", teamId).First(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to fetch team" + err.Error()})
		return
	}
	if team.OwnerId == userId || isAdmin {
		c.JSON(http.StatusOK, gin.H{"team": team})
	} else {
		c.JSON(http.StatusConflict, gin.H{"error": "only team creator and admin can access team "})
		return
	}

}
