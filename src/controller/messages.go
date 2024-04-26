package controllers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"minitwit.com/devops/logger"
	database "minitwit.com/devops/src/database"
	flash "minitwit.com/devops/src/flash"
	model "minitwit.com/devops/src/models"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func GetMessages(user string, page string, c *gin.Context) []map[string]interface{} {
	var results []map[string]interface{}
	logger.Log.Debug("Getting messages from: ",zap.String("User",user),zap.String("Page", page))
	user_query := c.Request.URL.Query().Get("username")

	offset, messagesPerPage := LimitMessages(page)

	userID := GetUser(user).ID

	if user == "" {
		database.DB.Table("messages").Select("messages.*, users.*").
			Joins("JOIN users ON messages.author = users.username").
			Where("messages.flagged = ?", false).
			Order("messages.created_at desc").
			Offset(offset).Limit(messagesPerPage).Find(&results)
	} else if user == user_query {
		database.DB.Table("messages").Where("author = ?", user).Limit(messagesPerPage).Order("created_at desc").Offset(offset).Find(&results)
	} else {
		database.DB.Table("messages").Select("messages.*, users.*").
			Joins("JOIN users ON messages.author = users.username").
			Where("(username = ? OR id IN (SELECT following FROM follows WHERE follower = ?)) AND messages.flagged = ?", user, userID, false).
			Order("messages.created_at desc").Offset(offset).Limit(messagesPerPage).Find(&results)
	}
	return results
}

func LimitMessages(page string) (int, int) {
	messagesPerPage := 50
	p, err := strconv.Atoi(page)
	if err != nil {
		panic("Failed to parse page number")
	}
	offset := (p - 1) * messagesPerPage
	return offset, messagesPerPage
}

func AddMessage(c *gin.Context) {
	user, err := c.Cookie("token")
	if err != nil || user == "" {
		logger.Log.Error("Failed to get token")
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	// Check if the user exists
	var count int64
	database.DB.Model(&model.User{}).Where("username = ?", user).Count(&count)
	if count == 0 {
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	// Check if the message is not empty
	message := c.PostForm("message")
	if strings.TrimSpace(message) == "" {
		c.Redirect(http.StatusFound, "/user_timeline")
		return
	}

	// Create and save the message
	t := time.Now()
	database.DB.Create(&model.Message{
		Author:    user,
		Text:      message,
		CreatedAt: t,
	})
	logger.Log.Info("Message recorded")
	flash.SetFlash(c, "message", "Your message was recorded")
	// Redirect to user timeline with a success message
	c.Redirect(http.StatusFound, "/user_timeline?message=success")
}

func GetFollower(follower uint, following uint) bool {
	var follows []model.Follow
	logger.Log.Info("Getting followers from: ", zap.Uint("follower",follower),zap.Uint("Following",following))
	if follower == following {
		return false
	} else {
		database.DB.Find(&follows).Where("follower = ?", following).Where("following = ?", follower).First(&follows)
		return len(follows) > 0
	}
}
