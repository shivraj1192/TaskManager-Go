package notificationController

import (
	"fmt"
	"task-manager/config"
	"task-manager/model"
)

func CreateNotifications(notifications []model.Notification) bool {
	DB := config.DB
	if err := DB.Create(&notifications).Error; err != nil {
		fmt.Println("@@@@@@@@@@@@" + err.Error())
		return false
	}
	return true
}
