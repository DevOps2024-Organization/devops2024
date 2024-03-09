package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/penglongli/gin-metrics/ginmetrics"
	controller "minitwit.com/devops/src/controller"
	database "minitwit.com/devops/src/database"
	model "minitwit.com/devops/src/models"
)

func getGinMetrics(router *gin.Engine) {
	// get global Monitor object
	m := ginmetrics.GetMonitor()
	// +optional set metric path, default /debug/metrics
	m.SetMetricPath("/ginmetrics")
	// +optional set slow time, default 5s
	m.SetSlowTime(10)
	// +optional set request duration, default {0.1, 0.3, 1.2, 5, 10}
	// used to p95, p99
	m.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})
	// set middleware for gin
	m.Use(router)
}

func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	hour, minute, second := t.Clock()
	return fmt.Sprintf("%02d/%02d/%d %02d:%02d:%02d", day, month, year, hour, minute, second)
}

func GetUserID(username string) uint {
	var user model.User
	database.DB.Where("username = ?", username).First(&user)
	return user.ID
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

	database.SetupDB()

	router := gin.Default()
	router.SetFuncMap(template.FuncMap{
		"formatAsDate": formatAsDate,
		"getUserId":    GetUserID,
	})
	router.LoadHTMLGlob("src/web/templates/*.tpl")
	router.Static("/web/static", "./src/web/static")

	router.GET("/", controller.Timeline)
	router.GET("/version", (func(c *gin.Context) {
		c.Data(200, "application/json; charset=utf-8", []byte(os.Getenv("VERSION")))
	}))
	router.GET("/public_timeline", controller.Timeline)
	router.GET("/user_timeline", controller.UserTimeline)
	router.GET("/register", controller.Register)
	router.POST("/register", controller.SignUp)
	router.GET("/login", controller.LoginPage)
	router.POST("/login", controller.Login)
	router.GET("/logout", controller.Logout)
	router.GET("/follow", controller.Follow)
	router.GET("/unfollow", controller.Unfollow)
	router.POST("/add_message", controller.AddMessage)

	getGinMetrics(router)

	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
