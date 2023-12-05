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
	session := sessions.Default(c)
	user := session.Get("user").(string)
	page := models.GetUserPage(user)
	populate(&page, user)
	projects, err := database.GetAllProjects()
	if err != nil {
		slog.Error(err.Error())
	} else {
		for _, project := range projects {
			page.Projects = append(page.Projects, project.Name)
		}
	}
	c.HTML(http.StatusOK, "layout", page)
}

func populate(page *models.Page, user string) {
	status, err := getStatus(user)
	if err != nil {
		log.Println("getStatus", err)
		return
	}
	page.Status = status
	page.Tracking = models.IsTrackingActive(user)
}
