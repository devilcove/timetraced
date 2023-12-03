package main

import (
	"embed"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

//go:embed: images/favicon.ico
var icon embed.FS

func setupRouter() *gin.Engine {
	//gin.SetMode(gin.ReleaseMode)
	secret, ok := os.LookupEnv("SESSION_SECRET")
	if !ok {
		secret = "secret"
	}
	store := cookie.NewStore([]byte(secret))
	session := sessions.Sessions("time", store)
	router := gin.Default()
	router.LoadHTMLGlob("html/*.html")
	router.Static("images", "./images")
	router.Static("assets", "./assets")
	router.StaticFS("/favicon.ico", http.FS(icon))
	//router.SetHTMLTemplate(template.Must(template.New("").Parse("html/*")))
	router.SetTrustedProxies(nil)
	router.Use(gin.Recovery(), session)
	users := router.Group("/users", auth)
	{
		users.GET("", getUsers)
		users.GET("current", getUser)
		users.POST("", addUser)
		users.PUT("", editUser)
		users.DELETE(":name", deleteUser)
	}
	router.GET("/login", displayLogin)
	router.POST("/login", login)
	router.GET("/logout", logout)
	router.GET("/register", register)
	router.POST("/register", regUser)
	router.GET("/configuration", config)
	router.POST("/setConfig", setConfig)
	status := router.Group("/", auth)
	{
		status.GET("/", displayStatus)
	}
	projects := router.Group("/projects", auth)
	{
		projects.GET("", getProjects)
		projects.POST("", addProject)
		projects.GET("/:name", getProject)
		projects.POST("/:name/start", start)
		projects.POST("/stop", stop)
		projects.GET("/status", displayStatus)
	}
	reports := router.Group("/reports", auth)
	{
		reports.GET("/", report)
		reports.POST("/", getReport)
	}
	return router
}

func processError(c *gin.Context, status int, message string) {
	slog.Error(message, "status", status)
	content := models.ErrorMessage{
		Status:  status,
		Message: message,
	}
	c.HTML(http.StatusOK, "error", content)
	c.Abort()
}

func auth(c *gin.Context) {
	session := sessions.Default(c)
	loggedIn := session.Get("loggedin")
	if loggedIn != true {
		models.SetPage("login")
		location := url.URL{Path: "/login"}
		c.Redirect(http.StatusFound, location.RequestURI())
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
		slog.Debug("user exists")
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
		slog.Error("hash error", "error", err)
	}
	database.SaveUser(&models.User{
		Username: user,
		Password: password,
		IsAdmin:  true,
		Updated:  time.Now(),
	})
	slog.Info("default user created")
}
