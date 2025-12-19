package main

import (
	"bytes"
	"crypto/rand"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/devilcove/mux"
	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/gorilla/sessions"
)

const (
	sessionBytes = 32
	cookieAge    = 300
)

var (
	store     *sessions.CookieStore
	templates *template.Template
	logger    *slog.Logger
)

// //go:embed images/favicon.ico
// var icon embed.FS

func setupRouter(l *slog.Logger) *mux.Router {
	store = sessions.NewCookieStore(randBytes(sessionBytes))
	store.MaxAge(cookieAge)
	store.Options.HttpOnly = true
	store.Options.SameSite = http.SameSiteStrictMode

	router := mux.NewRouter(l, mux.Logger)
	logger = l

	router.Static("/images/", "./images")
	router.Static("/assets/", "./assets")
	router.ServeFile("/favicon.ico", "images/favicon.ico")
	templates = template.Must(template.ParseGlob("html/*"))

	router.Post("/login", login)
	router.Get("/logout/", logout)
	router.Get("/{$}", displayMain)

	users := router.Group("/users", auth)
	users.Get("/{$}", getUsers)
	users.Get("/register/", register)
	users.Post("/register/", registerUser)
	users.Post("/{name}", editUser)
	users.Delete("/{name}", deleteUser)
	users.Get("/{name}", getUser)

	projects := router.Group("/projects", auth)
	projects.Get("/add/", displayProjectForm)
	projects.Post("/{$}", addProject)
	projects.Post("/stop/", stop)
	projects.Post("/start/{name}", start)

	reports := router.Group("/reports", auth)
	reports.Get("/{$}", report)
	reports.Post("/{$}", getReport)

	records := router.Group("/records", auth)
	records.Get("/{id}", getRecord)
	records.Post("/{id}", editRecord)

	configuration := router.Group("/config", auth)
	configuration.Get("/{$}", configOld)
	configuration.Post("/{$}", setConfig)
	return router
}

func processError(w http.ResponseWriter, status int, message string) {
	buf := bytes.Buffer{}
	l := log.New(&buf, "ERROR: ", log.Lshortfile)
	_ = l.Output(2, message)
	logger.Error(buf.String())
	http.Error(w, message, status)
}

func checkDefaultUser() {
	user := os.Getenv("USER")
	pass := os.Getenv("PASS")
	users, err := database.GetAllUsers()
	if err != nil {
		log.Fatal(err)
	}
	if len(users) > 0 {
		slog.Debug("user exists", "user", users[0].Username)
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
	_ = database.SaveUser(&models.User{
		Username: user,
		Password: password,
		IsAdmin:  true,
		Updated:  time.Now(),
	})
	logger.Info(
		"default user created",
		"user",
		user,
		"env user",
		os.Getenv("USER"),
		"env pass",
		os.Getenv("PASS"),
	)
}

func randBytes(l int) []byte {
	bytes := make([]byte, l)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return bytes
}
