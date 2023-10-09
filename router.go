package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	secret, ok := os.LookupEnv("SESSON_SECRET")
	if !ok {
		secret = "secret"
	}
	store := cookie.NewStore([]byte(secret))
	session := sessions.Sessions("time", store)
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(session)
	//router.POST("/newuser", New)
	router.POST("/login", login)
	router.GET("/logout", logout)
	projects := router.Group("/projects", auth)
	{
		projects.GET("", getProjects)
		projects.POST("", addProject)
		projects.GET("/:name", getProject)
		projects.POST("/start", start)
		projects.POST("/stop", stop)
		projects.GET("/status", status)
	}
	return router
}

func processError(c *gin.Context, status int, message string) {
	session := sessions.Default(c)
	session.Set("message", message)
	session.Save()
	c.JSON(status, gin.H{"message": message})
	c.Abort()
}

func auth(c *gin.Context) {
	session := sessions.Default(c)
	loggedIn := session.Get("loggedin")
	message := session.Get("message")
	user := session.Get("user")
	log.Println("auth", loggedIn, user, message)
	if loggedIn != true {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
		c.Abort()
		return
	}
}

func checkDefaultUser() {
	user := os.Getenv("user")
	pass := os.Getenv("pass")
	users, err := database.GetAllUsers()
	if err != nil {
		log.Fatal(err)
	}
	if len(users) > 1 {
		log.Println("user exists")
		return
	}
	if user == "" {
		user = "admin"
	}
	if pass == "" {
		pass = "password"
	}
	password, err := hashPassword(pass)
	if err != nil {
		log.Println("hash error", err)
	}
	database.SaveUser(&models.User{
		Username: user,
		Password: password,
		Updated:  time.Now(),
	})
	log.Println("default user created")
}
