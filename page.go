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
	page := models.GetPage()
	session := sessionData(r)
	page.NeedsLogin = true
	if session != nil {
		slog.Debug("displaying status for", "user", session.User, "loggedIn", session.LoggedIn)
		page = populatePage(session.User)
		if !session.LoggedIn {
			page.NeedsLogin = true
		}
	}
	slog.Debug(
		"displaystatus",
		"page",
		page.NeedsLogin,
		"refresh",
		page.Refresh,
		"theme",
		page.Theme,
	)
	_ = templates.ExecuteTemplate(w, "layout", page)
}

func displayStatus(w http.ResponseWriter, r *http.Request) {
	session := sessionData(r)
	user := session.User
	loggedIn := session.LoggedIn
	if user == "" || !loggedIn {
		displayMain(w, r)
		return
	}
	page := populatePage(user)
	_ = templates.ExecuteTemplate(w, "content", page)
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
