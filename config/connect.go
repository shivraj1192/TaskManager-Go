package config

import (
	"log"
	"os"
	"task-manager/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	dsn := "host=localhost user=" + user + " password=" + password + " dbname=taskmanager port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	DB = db
}

func CreateTables(DB *gorm.DB) {
	DB.AutoMigrate(&model.User{}, &model.Team{}, &model.Task{}, &model.Comment{}, &model.Notification{}, &model.Label{}, &model.Attachment{})
}

// ------------------------------- For Sqlite --------------------------------------

// package config

// import (
// 	"database/sql"
// 	"log"
// 	"task-manager/model"

// 	"gorm.io/driver/sqlite"
// 	"gorm.io/gorm"
// 	_ "modernc.org/sqlite"
// )

// var DB *gorm.DB

// func Connect() {
// 	dbsql, err := sql.Open("sqlite", "file:taskmanager.db?_busy_timeout=5000")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	var db *gorm.DB
// 	db, err = gorm.Open(sqlite.Dialector{Conn: dbsql}, &gorm.Config{})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	DB = db
// }

// func CreateTables(DB *gorm.DB) {
// 	DB.AutoMigrate(&model.User{}, &model.Team{}, &model.Task{}, &model.Comment{}, &model.Notification{}, &model.Label{}, &model.Attachment{})
// }
