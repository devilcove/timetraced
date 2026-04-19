package main

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/devilcove/cookie"
	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
)

func displayMain(w http.ResponseWriter, r *http.Request) {
	page := models.GetPage()
	user, err := cookie.Get(r, cookieName)
	if err != nil {
		render(w, "login", page)
		return
	}
	page = populatePage(string(user))
	render(w, "layout", page)
}

func displayStatus(w http.ResponseWriter, r *http.Request) {
	user, err := cookie.Get(r, cookieName)
	if err != nil {
		render(w, "loginForm", nil)
		return
	}
	page := populatePage(string(user))
	render(w, "content", page)
}

func populatePage(user string) models.Page {
	page := models.GetUserPage(user)
	page.Tracking = models.IsTrackingActive(user)
	projects, err := database.GetAllProjects()
	if err != nil {
		slog.Error(err.Error())
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
