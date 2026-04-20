package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/devilcove/cookie"
	"github.com/devilcove/mux"
	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
)

const (
	cookieAge  = 300
	cookieName = "devilcove-time"
)

var templates *template.Template

// //go:embed images/favicon.ico
// var icon embed.FS

func setupRouter() *mux.Router {
	if err := cookie.New(cookieName, cookieAge); err != nil {
		log.Fatal("set cookie", err)
	}
	// store = sessions.NewCookieStore(randBytes(sessionBytes))
	// store.MaxAge(cookieAge)
	// store.Options.HttpOnly = tru)e
	// store.Options.SameSite = http.SameSiteStrictMode

	router := mux.NewRouter(mux.Logger)

	router.Static("/images/", "./images")
	router.Static("/assets/", "./assets")
	router.ServeFile("/favicon.ico", "images/favicon.ico")
	templates = template.Must(template.ParseGlob("html/*"))

	router.Post("/login", login)
	router.Get("/logout/", logout)
	router.Get("/{$}", displayMain)

	status := router.Group("/status", auth)
	status.Get("/{$}", displayStatus)

	users := router.Group("/users", auth)
	users.Get("/{$}", getUsers)
	users.Get("/register/", register)
	users.Post("/register/", registerUser)
	users.Post("/{name}", editUser)
	users.Delete("/{name}", deleteUser)
	users.Get("/{name}", getUser)

	projects := router.Group("/projects", auth)
	projects.Get("/list/", showProjects)
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
	slog.Error(buf.String())
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
	slog.Info(
		"default user created",
		"user",
		user,
		"env user",
		os.Getenv("USER"),
		"env pass",
		os.Getenv("PASS"),
	)
}

func render(w io.Writer, template string, data any) {
	if err := templates.ExecuteTemplate(w, template, data); err != nil {
		slog.Error("render template", "caller", caller(2), "name", template,
			"data", data, "error", err)
	}
}

func caller(depth int) string {
	pc, file, no, ok := runtime.Caller(depth)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		return fmt.Sprintf("%s %s:%d", details.Name(), filepath.Base(file), no)
	}
	return "unknown caller"
}
