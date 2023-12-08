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

func displayStatus(c *gin.Context) {
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
	if loggedIn != nil {
		page.DisplayLogin = !loggedIn.(bool)
	}
	slog.Info("displaystatus", "page", loggedIn)
	c.HTML(http.StatusOK, "layout", page)
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
