package main

import (
	"log"
	"log/slog"
	"net/http"
	"time"

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
	slog.Debug("displaying status for", "user", user, "loggedIn", loggedIn)
	if user == nil {
		page = populatePage("")
	} else {
		page = populatePage(user.(string))
	}
	if loggedIn == nil {
		page.NeedsLogin = true
	}
	slog.Debug("displaystatus", "page", page.NeedsLogin, "refresh", page.Refresh)
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
	page.DefaultDate = time.Now().Local().Format("2006-01-02")
	return page
}
