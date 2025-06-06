package main

import (
	"os"
	"task-manager/config"
	"task-manager/routes"

	"github.com/gin-gonic/gin"
	"github.com/subosito/gotenv"
)

func init() {
	gotenv.Load("../.env")
	config.Connect()
	config.CreateTables(config.DB)
}

func main() {
	router := gin.Default()
	routes.SetUpRoutes(router)
	port := os.Getenv("PORT")
	router.Run(":" + port)
}
