package main

import (
	"fmt"
	"log/slog"
	"net/http"
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
		processError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, projects)
}

func displayProjectForm(c *gin.Context) {
	c.HTML(http.StatusOK, "addProject", "")
}

func addProject(c *gin.Context) {
	var project models.Project
	if err := c.Bind(&project); err != nil {
		processError(c, http.StatusBadRequest, "could not decode request into json "+err.Error())
		return
	}
	slog.Info("addproject1", "project", project)
	if regexp.MustCompile(`\s+`).MatchString(project.Name) || project.Name == "" {
		processError(c, http.StatusBadRequest, "invalid project name")
		return
	}
	existing, err := database.GetProject(project.Name)
	if err != nil {
		processError(c, http.StatusInternalServerError, "database error")
		return
	}
	if existing.Name == project.Name {
		processError(c, http.StatusBadRequest, "project exists")
		return
	}
	project.ID = uuid.New()
	project.Active = true
	project.Updated = time.Now()
	slog.Info("add project", "project", project)
	if err := database.SaveProject(&project); err != nil {
		processError(c, http.StatusInternalServerError, "error saving project "+err.Error())
		return
	}
	slog.Info("added", "project", project.Name)
	displayStatus(c)
}

func getProject(c *gin.Context) {
	p := c.Param("name")
	project, err := database.GetProject(p)
	if err != nil {
		processError(c, http.StatusBadRequest, "could not retrieve project "+err.Error())
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
		processError(c, http.StatusInternalServerError, "error reading project "+err.Error())
		return
	}
	if !project.Active {
		processError(c, http.StatusBadRequest, "project is not active")
		return
	}
	if models.IsTrackingActive(user) {
		if err := stopE(user); err != nil {
			processError(c, http.StatusInternalServerError, err.Error())
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
		processError(c, http.StatusInternalServerError, "failed to save record "+err.Error())
		return
	}
	models.TrackingActive(user, project)
	slog.Info("tracking started", "project", project.Name)
	displayStatus(c)
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
		processError(c, http.StatusInternalServerError, err.Error())
		return
	}
	displayStatus(c)
}
