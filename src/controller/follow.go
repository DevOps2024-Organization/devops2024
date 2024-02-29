package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	database "minitwit.com/devops/src/database"
	flash "minitwit.com/devops/src/flash"
	model "minitwit.com/devops/src/models"
)

func Follow(c *gin.Context) {
	user_to_follow := c.Request.URL.Query().Get("username")
	user, _ := c.Cookie("token")
	if user == "" {
		panic("You must be logged in to follow users")
	} else {
		database.DB.Create(&model.Follow{Follower: GetUser(user).ID, Following: GetUser(user_to_follow).ID})
	}
	flash.SetFlash(c, "message", fmt.Sprintf("You are now following %s", user_to_follow))
	c.Redirect(http.StatusFound, "/user_timeline?username="+user_to_follow)
}

func Unfollow(c *gin.Context) {
	var follows []model.Follow
	user_to_follow := c.Request.URL.Query().Get("username")
	user, _ := c.Cookie("token")
	if user == "" {
		panic("You must be logged in to follow users.")
	} else {
		database.DB.Where("follower = ?", GetUser(user).ID).Where("following = ?", GetUser(user_to_follow).ID).Delete(&follows)

	}
	flash.SetFlash(c, "message", fmt.Sprintf("You are no longer following %s", user_to_follow))
	c.Redirect(http.StatusFound, "/user_timeline?username="+user_to_follow)
}
