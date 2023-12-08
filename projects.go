package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func getProjects(c *gin.Context) {
	projects, err := database.GetAllProjects()
	if err != nil {
		processError(c, "ServerError", err.Error())
		return
	}
	c.JSON(http.StatusOK, projects)
}

func addProject(c *gin.Context) {
	var project models.Project
	if err := c.BindJSON(&project); err != nil {
		processError(c, "BadRequest", "could not decode request into json "+err.Error())
		return
	}
	if regexp.MustCompile(`\s+`).MatchString(project.Name) {
		processError(c, "BadRequest", "invalid project name")
		return
	}
	if _, err := database.GetProject(project.Name); err == nil {
		processError(c, "BadRequest", "project exists")
		return
	}
	project.ID = uuid.New()
	project.Active = true
	project.Updated = time.Now()
	if err := database.SaveProject(&project); err != nil {
		processError(c, "ServerError", "error saving project "+err.Error())
		return
	}
	slog.Info("added", "project", project.Name)
	location := url.URL{Path: "/"}
	c.Redirect(http.StatusFound, location.RequestURI())
}

func getProject(c *gin.Context) {
	p := c.Param("name")
	project, err := database.GetProject(p)
	if err != nil {
		processError(c, "BadRequest", "could not retrieve project "+err.Error())
		return
	}
	c.JSON(http.StatusOK, project)
}

func start(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user").(string)
	p := c.Param("name")
	project, err := database.GetProject(p)
	if err != nil {
		processError(c, "ServerError", "error reading project "+err.Error())
		return
	}
	if !project.Active {
		processError(c, "BadRequest", "project is not active")
		return
	}
	if models.IsTrackingActive(user) {
		if err := stopE(user); err != nil {
			processError(c, "ServerError", err.Error())
			return
		}
	}
	record := models.Record{
		ID:      uuid.New(),
		Project: p,
		User:    user,
		Start:   time.Now(),
	}
	if err := database.SaveRecord(&record); err != nil {
		processError(c, "ServerError", "failed to save record "+err.Error())
		return
	}
	models.TrackingActive(user, project)
	slog.Info("tracking started", "project", project.Name)
	location := url.URL{Path: "/"}
	c.Redirect(http.StatusFound, location.RequestURI())
}

func stopE(u string) error {
	records, err := database.GetAllRecordsForUser(u)
	if err != nil {
		return fmt.Errorf("failed to retrieve records %w", err)
	}
	for _, record := range records {
		if record.End.IsZero() {
			record.End = time.Now()
			if err := database.SaveRecord(&record); err != nil {
				slog.Error("failed to save updated record", "error", err)
			}
		}
	}
	slog.Info("tracking stopped", "project", models.Tracked(u))
	models.TrackingInactive(u)
	return nil
}

func stop(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user").(string)
	if err := stopE(user); err != nil {
		processError(c, "ServerError", err.Error())
		return
	}
	location := url.URL{Path: "/"}
	c.Redirect(http.StatusFound, location.RequestURI())
}
