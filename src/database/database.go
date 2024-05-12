package controllers

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"minitwit.com/devops/logger"
	model "minitwit.com/devops/src/models"
)

var DB *gorm.DB

func SetupDB() {
	logger.Log.Info("Setup DB, getting env variables")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_DATABASE"),
		os.Getenv("DB_PORT"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal("failed to connect to database")
		fmt.Printf("Error: %s", err.Error())
		panic("Failed to connect to database.")
	}
	db.AutoMigrate(&model.User{}, &model.Message{}, &model.Follow{})
	DB = db
}
