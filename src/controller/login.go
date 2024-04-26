package controllers

import (
	"net/http"
	"strings"

	"minitwit.com/devops/src/flash"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"minitwit.com/devops/logger"
	database "minitwit.com/devops/src/database"
	model "minitwit.com/devops/src/models"
)

func GetUser(username string) model.User {
	var user model.User
	logger.Log.Debug("Getting ",zap.String("Username",username))
	database.DB.Where("username = ?", username).First(&user)
	return user
}

func PasswordCompare(salt string, password string, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(salt+password))
	return err
}

func ValidUser(username string, password string) (bool, string) {

	user := GetUser(username)

	if user.Username == "" {
		return false, "Invalid username"
	}

	err := PasswordCompare(user.Salt, password, user.Password)
	if err != nil {
		logger.Log.Error("Error occurred while validating password with error message")
		return false, "Invalid password"
	}

	return true, ""
}

func Login(c *gin.Context) {
	logger.Log.Info("Loggin in...")
	username := strings.ToLower(c.Request.FormValue("username"))
	password := c.Request.FormValue("password")

	if strings.Trim(username, " ") == "" || strings.Trim(password, " ") == "" {
		c.HTML(http.StatusOK, "login.tpl", gin.H{
			"ErrorTitle":   "Empty Fields",
			"ErrorMessage": "Please fill in all fields",
		})
	}

	valid, errMsg := ValidUser(username, password)
	if valid {
		flash.SetFlash(c, "message", "You were logged in")
		c.SetCookie("token", username, 3600, "", "", false, true)
		c.Redirect(http.StatusFound, "/user_timeline")
	} else {
		logger.Log.Error("Login invalid with ",zap.String("Error message",errMsg))
		// Send back the specific error message (errMsg) in the response.
		c.HTML(http.StatusOK, "login.tpl", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": errMsg,
		})
	}
}

func LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.tpl", gin.H{
		"title": "Login",
	})
}

func Logout(c *gin.Context) {
	logger.Log.Info("Logging out...")
	flash.SetFlash(c, "message", "You were logged out")
	c.SetCookie("token", "", -1, "", "", false, true)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}
