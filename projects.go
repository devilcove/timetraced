package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
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

func addProject(c *gin.Context) {
	var project models.Project
	if err := c.BindJSON(&project); err != nil {
		processError(c, http.StatusBadRequest, "could not decode request into json "+err.Error())
		return
	}
	if _, err := database.GetProject(project.Name); err == nil {
		processError(c, http.StatusBadRequest, "project exists ")
		return
	}
	project.ID = uuid.New()
	project.Active = true
	project.Updated = time.Now()
	if err := database.SaveProject(&project); err != nil {
		processError(c, http.StatusInternalServerError, "error saving project "+err.Error())
		return
	}
	slog.Info("added", "project", project.Name)
	c.JSON(http.StatusOK, project)
}
func getProject(c *gin.Context) {
	p := c.Param("name")
	project, err := database.GetProject(p)
	if err != nil {
		processError(c, http.StatusBadGateway, "could not retrieve project "+err.Error())
		return
	}
	c.JSON(http.StatusOK, project)
}

func start(c *gin.Context) {
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
	if models.IsTrackingActive() {
		if err := stopE(); err != nil {
			processError(c, http.StatusInternalServerError, err.Error())
		}
	}
	record := models.Record{
		ID:      uuid.New(),
		Project: p,
		Start:   time.Now(),
	}
	if err := database.SaveRecord(&record); err != nil {
		processError(c, http.StatusInternalServerError, "failed to save record "+err.Error())
		return
	}
	models.TrackingActive(project)
	slog.Info("tracking started", "project", project.Name)
	c.JSON(http.StatusOK, project)
}

func stopE() error {
	records, err := database.GetAllRecords()
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
	slog.Info("tracking stopped", "project", models.Tracked())
	models.TrackingInactive()
	return nil
}

func stop(c *gin.Context) {
	if err := stopE(); err != nil {
		processError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, nil)
}

func status(c *gin.Context) {
	status, err := getStatus()
	if err != nil {
		processError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, status)
}
