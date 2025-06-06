package notificationController

import (
	"fmt"
	"task-manager/config"
	"task-manager/model"
)

func CreateNotification(notification model.Notification) bool {
	fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@ start @@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	DB := config.DB

	if err := DB.Create(&notification).Error; err != nil {
		fmt.Println("@@@@@@@@@@@@" + err.Error())
		return false
	}
	return true
}
