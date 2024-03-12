package main

import (
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"
	"flag"
	// "encoding/json"

  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promauto"
  "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	model "minitwit.com/devops/src/models"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "api_http_requests_total",
        Help: "Total number of HTTP requests.",
    }, []string{"method", "endpoint", "status_code"})

    requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "api_request_duration_seconds",
        Help:    "Duration of HTTP requests in seconds.",
        Buckets: prometheus.DefBuckets,
    }, []string{"method", "endpoint"})

    inFlightRequests = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "api_in_flight_requests",
        Help: "Current number of in-flight requests.",
    })
)

func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Increment in-flight requests gauge
        inFlightRequests.Inc()

        start := time.Now()
        c.Next() // Process request
        duration := time.Since(start)

        // Decrement in-flight requests gauge
        inFlightRequests.Dec()

        status := strconv.Itoa(c.Writer.Status())
        endpoint := c.Request.URL.Path // Or use c.FullPath() for matching route
        method := c.Request.Method

        httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
        requestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
    }
}

type APIMessage struct {
	User      string `json:"user"`
	CreatedAt string `json:"created_at"`
	Flagged   bool   `json:"flagged"`
	MessageID int    `json:"message_id"`
	Content   string `json:"content"`
}

type FollowerListStruct struct {
	Follows []string `json:"follows"`
}

var DB *gorm.DB
var LATEST = 0

func CreateUser(username string, email string, password string) bool {
	salt := Salt()
	err := DB.Create(&model.User{Username: username, Email: email, Salt: salt, Password: Hash(salt + password)}).Error
	if err != nil {
		return false
	} else {
		return true
	}
}

func Salt() string {
	bytes, _ := bcrypt.GenerateFromPassword(make([]byte, 8), 8)
	return string(bytes)
}

func Hash(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}

func SetupDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_DATABASE"),
		os.Getenv("DB_PORT"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		panic("Failed to connect to ")
	}
	db.AutoMigrate(&model.User{}, &model.Message{}, &model.Follow{})
	DB = db
}

func ValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func ValidRegistration(c *gin.Context, username string, email string, password1 string) bool {
	if username == "" {
		c.JSON(400, gin.H{
			"error": "You have to enter a username",
		})
		return false
	}
	if password1 == "" {
		c.JSON(400, gin.H{
			"error": "You have to enter a password",
		})
		return false
	}
	if !ValidEmail(email) {
		c.JSON(400, gin.H{
			"error": "You have to enter a valid email address",
		})
		return false
	}

	return true
}

func ConvertToAPIMessage(messages []model.Message) []APIMessage {
	var apiMessages []APIMessage

	for _, msg := range messages {
		apiMessage := APIMessage {
			User:      msg.Author,
			CreatedAt: msg.CreatedAt.Format(time.RFC3339),
			Flagged:   msg.Flagged,
			MessageID: int(msg.MessageID),
			Content:   msg.Text,
		}
		apiMessages = append(apiMessages, apiMessage)
	}
	return apiMessages
}

func SignUp(c *gin.Context) {
	//If directories are unreferenced then they should be removed from the web root and/or the application directory.
	// reponse: HTTP/1.1 301 Moved Permanently

	Latest(c)
	var json model.RegisterForm

	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// We have auto-complete here
	//Should maybe be fixed in the html: The first and most secure location is to disable the autocomplete attribute on the <form> HTML tag.

	username := json.Username
	email := json.Email
	password := json.Password

	if !ValidRegistration(c, username, email, password) {
		return
	}

	fmt.Println(username)
	if CreateUser(username, email, password) {
		c.JSON(204, gin.H{})
	} else {
		c.JSON(400, gin.H{"error": "The username is already taken"})
	}
}

// PasswordCompare handles password hash compare
func PasswordCompare(salt string, password string, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(salt+password))
	return err
}

func ValidUser(username string, password string) bool {

	user := GetUser(username)

	if user.Username == "" {
		return false
	}

	err := PasswordCompare(user.Salt, password, user.Password)

	return err == nil
}

func GetUser(username string) model.User {
	var user model.User
	DB.Where("username = ?", username).First(&user)
	return user
}

func GetMessages(user string, no int) []APIMessage {
	var messages []model.Message

	query := DB.Table("messages").Order("created_at desc").Limit(no)
	if user != "" {
		query = query.Where("author = ?", user)
	}
	query.Find(&messages)

	apiMessages := ConvertToAPIMessage(messages)
	return apiMessages
}

func GetFollowers(user uint) []string {
	var usernames []string
	err := DB.Model(&model.User{}).
		Select("users.username").
		Joins("JOIN follows ON follows.following = users.id").
		Where("follows.follower = ?", user).
		Pluck("users.username", &usernames).Error

	if err != nil {
		log.Printf("Error finding followers: %v", err)
		return nil
	}


	return usernames
}

func GetFollower(follower uint, following uint) bool {
	var follows []model.Follow
	if follower == following {
		return false
	} else {
		DB.Find(&follows).Where("follower = ?", following).Where("following = ?", follower).First(&follows)
		return len(follows) > 0
	}
}

func Follow(followerID uint, followeeID uint) *gorm.DB {
	follow := model.Follow{Follower: followerID, Following: followeeID}
	err := DB.Create(&follow)
	return err
}

func Unfollow(follower uint, followee uint) *gorm.DB {
	err := DB.Delete(&model.Follow{}, model.Follow{Follower: follower, Following: followee})
	return err
}

func sanitize(s string) string {
	return strings.ToValidUTF8(s, "")
}

func AddMessage(user string, message string) {
	//fix cross site forgery here
	//{
	// r := gin.Default()
	// store := cookie.NewStore([]byte("secret"))
	// r.Use(sessions.Sessions("mysession", store))
	// r.Use(csrf.Middleware(csrf.Options{
	// 	Secret: "secret123",
	// 	ErrorFunc: func(c *gin.Context) {
	// 		c.String(400, "CSRF token mismatch")
	// 		c.Abort()
	// 	},
	// }))
	message = sanitize(message)
	t := time.Now().Format(time.RFC822)
	time_now, _ := time.Parse(time.RFC822, t)
	DB.Create(&model.Message{Author: user, Text: message, CreatedAt: time_now})
}

func Latest(c *gin.Context) {
	l := c.Request.URL.Query().Get("latest")
	if l == "" {
		if c.FullPath() == "/latest" {
			c.JSON(200, gin.H{"latest": LATEST})
		}
		return
	}
	latest, err := strconv.Atoi(l)
	if err != nil {
		c.JSON(400, gin.H{"error_msg": "Latest must be an integer"})
		return
	}
	LATEST = latest
	if c.FullPath() == "/latest" {
		c.JSON(200, gin.H{"latest": LATEST})
	}

}

func main() {
	var isTest bool
	flag.BoolVar(&isTest,"test",false,"Set true if is test")
	flag.Parse()
	var envPath string = ".env"
	if isTest  {
		envPath = ".env-test"
	}
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading .env file")
	}

	SetupDB()

	router := gin.Default()
  // Register Prometheus middleware
  router.Use(PrometheusMiddleware())

  // Expose the metrics endpoint
  router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	//API ENDPOINTS ADDED
	router.GET("/", (func(c *gin.Context) {
		Latest(c)
		c.JSON(200, "Welcome to Go MiniTwit API!")
	}))

	router.GET("/version", (func(c *gin.Context) {
		Latest(c)
		c.Data(200, "application/json; charset=utf-8", []byte(os.Getenv("VERSION")))
	}))

	router.POST("/register", SignUp)

	// /msgs/*param means that param is optional
	// /msgs/:param means that param is required
	router.GET("/msgs/*usr", (func(c *gin.Context) {
		Latest(c)
		user := strings.Trim(c.Param("usr"), "/")
		no, err := strconv.Atoi(c.Request.URL.Query().Get("no"))
		if err != nil {
			no = 100
		}
		var data []APIMessage
		if user == "" {
			data = GetMessages("", no)
		} else {
			data = GetMessages(user, no)
		}
		if len(data) == 0 {
			c.Status(http.StatusNoContent)
		} else {
			c.JSON(http.StatusOK, data)
		}
	}))
	router.POST("/msgs/:usr", (func(c *gin.Context) {
		Latest(c)
		user := strings.Trim(c.Param("usr"), "/")

		if GetUser(user).ID == 0 {
			c.JSON(404, gin.H{"error": "user not found"})
			return
		}

		var message model.MessageForm

		if err := c.ShouldBindJSON(&message); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error_msg": "You must provide a message"})
			return
		}

		AddMessage(user, message.Content)
		c.JSON(http.StatusNoContent, gin.H{})
	}))

	router.GET("/latest", Latest)

	router.GET("/fllws/:usr", (func(c *gin.Context) {
		Latest(c)
		user := GetUser(strings.Trim(c.Param("usr"), "/"))
		if user.ID == 0 {
			c.JSON(404, gin.H{"error": "user not found"})
			return
		} else {
			followerList := GetFollowers(user.ID)
			followerListResponse := FollowerListStruct{Follows: followerList}
			c.JSON(http.StatusOK, followerListResponse)
			return
		}
	}))

	router.POST("/fllws/:usr", (func(c *gin.Context) {
		Latest(c)
		user := GetUser(strings.Trim(c.Param("usr"), "/"))
		if user.ID == 0 {
			c.JSON(404, gin.H{"error": "user not found"})
			return
		}

		var follow model.FollowForm
		if err := c.ShouldBindJSON(&follow); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if follow.Follow != "" {
			followee := GetUser(follow.Follow)
			err := Follow(user.ID, followee.ID)
			if err.Error != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": ""})
				return
			}
			c.JSON(http.StatusNoContent, gin.H{})
			return
		} else if follow.Unfollow != "" {
			unfollowee := GetUser(follow.Unfollow)
			err := Unfollow(user.ID, unfollowee.ID)
			if err.Error != nil {
				c.JSON(403, gin.H{"error": ""})
				return
			}
			c.Status(http.StatusNoContent)
			return
		} else if len(follow.Latest) > 0 {
			latest, err := strconv.Atoi(follow.Latest[0])
			if err != nil {
				c.JSON(403, gin.H{"error_msg": "Latest must be an integer"})
				return
			}
			LATEST = latest
			c.JSON(http.StatusOK, gin.H{"data": GetFollowers(user.ID)})
		} else {
			c.JSON(403, gin.H{"error_msg": "Only these fields are accepted: follow | unfollow | latest"})
			return
		}

	}))

	err := router.Run(":5000")
	if err != nil {
		panic(err)
	}
}
