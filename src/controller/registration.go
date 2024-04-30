package controllers

import (
	"net/http"
	"net/mail"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"minitwit.com/devops/logger"
	database "minitwit.com/devops/src/database"
	flash "minitwit.com/devops/src/flash"
	model "minitwit.com/devops/src/models"
)

func CreateUser(username string, email string, password string) {
	salt := Salt()
	usr := strings.ToLower(username)
	logger.Log.Debug("Creating user", zap.String("User", usr), zap.String("Email", email))
	database.DB.Create(&model.User{Username: usr, Email: email, Salt: salt, Password: Hash(salt + password)})
}

func Salt() string {
	bytes, _ := bcrypt.GenerateFromPassword(make([]byte, 8), 8)
	return string(bytes)
}

func Hash(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}

func ValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func ValidRegistration(c *gin.Context, username string, email string, password1 string, password2 string) bool {
	if password1 != password2 {
		logger.Log.Info("Password mismatch")
		c.HTML(http.StatusOK, "register.tpl", gin.H{
			"error": "The two passwords do not match",
		})
		return false
	}
	if username == "" {
		logger.Log.Info("Expecting username input")
		c.HTML(http.StatusOK, "register.tpl", gin.H{
			"error": "You have to enter a username",
		})
		return false
	}
	if password1 == "" {
		logger.Log.Info("Expecting password input")
		c.HTML(http.StatusOK, "register.tpl", gin.H{
			"error": "You have to enter a password",
		})
		return false
	}
	if !ValidEmail(email) {
		logger.Log.Info("Invalid email")
		c.HTML(http.StatusOK, "register.tpl", gin.H{
			"error": "You have to enter a valid email address",
		})
		return false
	}

	return true
}

func Register(c *gin.Context) {
	c.HTML(http.StatusOK, "register.tpl", gin.H{
		"title": "Register",
	})
}

func SignUp(c *gin.Context) {
	c.Request.ParseForm()
	username := c.Request.PostForm.Get("username")
	email := c.Request.PostForm.Get("email")
	password1 := c.Request.PostForm.Get("password1")
	password2 := c.Request.PostForm.Get("password2")

	if !ValidRegistration(c, username, email, password1, password2) {
		logger.Log.Info("Registration failed")
		return
	}

	var user model.User
	result := database.DB.Where("username = ?", strings.ToLower(username)).First(&user)
	if result.RowsAffected > 0 {
		logger.Log.Info("User already exists", zap.String("Username", username))
		c.HTML(http.StatusOK, "register.tpl", gin.H{
			"error": "The username is already taken",
		})
		return
	}

	CreateUser(username, email, password1)
	location := url.URL{Path: "/login"}
	flash.SetFlash(c, "message", "You were successfully registered and can login now")
	data := make(map[string]interface{})
	data["flashes"] = flash.GetFlash(c, "message")
	c.HTML(http.StatusOK, "register.tpl", gin.H{
		"flashes": data,
	})
	c.Redirect(http.StatusFound, location.RequestURI())
}
