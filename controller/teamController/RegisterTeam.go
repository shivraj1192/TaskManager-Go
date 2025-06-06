package teamController

import (
	"net/http"
	"task-manager/config"
	"task-manager/controller/notificationController"
	"task-manager/model"

	"github.com/gin-gonic/gin"
)

func RegisterTeam(c *gin.Context) {
	var team model.Team
	var user model.User
	DB := config.DB

	// fmt.Println("Start")

	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get user ID from context"})
		return
	}

	// fmt.Println("Start1")

	if err := c.ShouldBindJSON(&team); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team input: " + err.Error()})
		return
	}

	// fmt.Println(userID)

	if err := DB.First(&user, userID.(uint)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found: " + err.Error()})
		return
	}
	// fmt.Println("Start2")

	// Team register notification
	notification := model.Notification{
		Content: "You have created " + team.Name + " team.",
		UserID:  user.ID,
	}

	if ok := notificationController.CreateNotification(notification); !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	team.OwnerId = userID.(uint)

	if err := DB.Create(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team: " + err.Error()})
		return
	}

	// fmt.Println("Start3")

	if err := DB.Model(&team).Association("Members").Append(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add owner to team members: " + err.Error()})
		return
	}

	// fmt.Println("Start4")

	if err := DB.Preload("Members").First(&team, team.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load team members: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"team": team})
}
