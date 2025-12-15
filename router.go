package main

import (
	"crypto/rand"
	"embed"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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

//go:embed images/favicon.ico
var icon embed.FS

func setupRouter(l *slog.Logger) *mux.Router { //nolint:funlen
	store = sessions.NewCookieStore(randBytes(sessionBytes))
	store.MaxAge(cookieAge)
	store.Options.HttpOnly = true
	store.Options.SameSite = http.SameSiteStrictMode

	router := mux.NewRouter(l, mux.Logger)
	logger = l

	// router.LoadHTMLGlob("html/*.html")
	router.Static("/images/", "./images")
	router.Static("/assets/", "./assets")
	router.ServeFile("/favicon.ico", "images/favicon.ico")
	templates = template.Must(template.ParseGlob("html/*"))
	// templates.Lookup("about").Execute(os.Stdout, nil)
	// router.SetHTMLTemplate(template.Must(template.New("").Parse("html/*")))
	// _ = router.SetTrustedProxies(nil)
	// router.Use(gin.Recovery(), session)
	users := router.Group("/users", auth)
	users.Get("/{$}", getUsers)
	users.Get("/register/", register)
	users.Post("/register/", registerUser)
	//users.Get("current", getUser)
	// 	users.Post("", addUser)
	users.Post("/{name}", editUser)
	users.Delete("/{name}", deleteUser)
	users.Get("/{name}", getUser)
	// }
	// router.Get("/login", displayLogin)
	router.Post("/login", login)
	router.Get("/logout/", logout)
	router.Get("/{$}", displayMain)
	// router.Post("/register", regUser)
	projects := router.Group("/projects", auth)
	//projects.Get("/{$}", getProjects)
	projects.Get("/add/", displayProjectForm)
	projects.Post("/{$}", addProject)
	// 	projects.Get("/:name", getProject)
	projects.Post("/stop/", stop)
	projects.Post("/start/{name}", start)
	// 	projects.Get("/status", displayStatus)
	// }
	reports := router.Group("/reports", auth)
	// {
	reports.Get("/{$}", report)
	reports.Post("/{$}", getReport)
	// }
	records := router.Group("records", auth)
	// {
	records.Get("/{id}", getRecord)
	records.Post("/{id}", editRecord)
	// }
	configuration := router.Group("/config", auth)
	// {
	configuration.Get("/{$}", configOld)
	configuration.Post("/{$}", setConfig)
	// }
	return router
}

func processError(w http.ResponseWriter, status int, message string) {
	pc, fn, line, _ := runtime.Caller(1)
	source := fmt.Sprintf("%s[%s:%d]", runtime.FuncForPC(pc).Name(), filepath.Base(fn), line)
	slog.Error(message, "status", status, "source", source)
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
	slog.Info("default user created", "user", user)
}

func randBytes(l int) []byte {
	bytes := make([]byte, l)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return bytes
}
