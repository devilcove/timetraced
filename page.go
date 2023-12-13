package main

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func displayMain(c *gin.Context) {
	var page models.Page
	session := sessions.Default(c)
	user := session.Get("user")
	loggedIn := session.Get("loggedin")
	slog.Info("displaying status for", "user", user, "loggedIn", loggedIn)
	if user == nil {
		page = populatePage("")
	} else {
		page = populatePage(user.(string))
	}
	if loggedIn == nil {
		page.NeedsLogin = true
		slog.Info("setting needs login", "needsLogin", page.NeedsLogin)
	}
	slog.Info("displaystatus", "page", page.NeedsLogin)
	c.HTML(http.StatusOK, "layout", page)
}

func displayStatus(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	loggedIn := session.Get("loggedin")
	if user == nil || loggedIn == nil {
		displayMain(c)
		return
	}
	page := populatePage(user.(string))
	c.HTML(http.StatusOK, "content", page)
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
	return page
}
