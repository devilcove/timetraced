package main

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
)

func displayMain(w http.ResponseWriter, r *http.Request) {
	user := getRequestUser(r)
	page := populatePage(user.Username)
	slog.Info("main page", "user", user.Username)
	render(w, "layout", page)
}

func displayStatus(w http.ResponseWriter, r *http.Request) {
	user := getRequestUser(r)
	page := populatePage(user.Username)
	render(w, "content", page)
}

func populatePage(user string) models.Page {
	page := models.GetPage()
	page.Tracking = models.IsTrackingActive(user)
	projects, err := database.GetAllProjects()
	if err != nil {
		slog.Error("get projects", "error", err)
	} else {
		for _, project := range projects {
			page.Projects = append(page.Projects, project.Name)
		}
	}
	status, err := getStatus(user)
	if err != nil {
		log.Println("getStatus", err)
	}
	page.Status = status
	page.DefaultDate = time.Now().Local().Format("2006-01-02")
	return page
}
