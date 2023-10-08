package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetProjects(c *gin.Context) {
	projects, err := database.GetAllProjects()
	if err != nil {
		ProcessError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, projects)
}

func AddProject(c *gin.Context) {
	var project models.Project
	if err := c.BindJSON(&project); err != nil {
		ProcessError(c, http.StatusBadRequest, "could not decode request into json "+err.Error())
		return
	}
	if _, err := database.GetProject(project.Name); err == nil {
		ProcessError(c, http.StatusBadRequest, "project exists ")
		return
	}
	project.ID = uuid.New()
	project.Active = true
	project.Updated = time.Now()
	if err := database.SaveProject(&project); err != nil {
		ProcessError(c, http.StatusInternalServerError, "error saving project "+err.Error())
		return
	}
	c.JSON(http.StatusOK, project)
}
func GetProject(c *gin.Context) {
	p := c.Param("name")
	project, err := database.GetProject(p)
	if err != nil {
		ProcessError(c, http.StatusBadGateway, "could not retrieve project "+err.Error())
		return
	}
	c.JSON(http.StatusOK, project)
}

func Start(c *gin.Context) {
	p := c.Param("name")
	project, err := database.GetProject(p)
	if err != nil {
		ProcessError(c, http.StatusInternalServerError, "error reading project "+err.Error())
		return
	}
	if !project.Active {
		ProcessError(c, http.StatusBadRequest, "project is not active")
		return
	}
	if models.IsTrackingActive() {
		if err := stop(); err != nil {
			ProcessError(c, http.StatusInternalServerError, err.Error())
		}
	}
	record := models.Record{
		ID:      uuid.New(),
		Project: p,
		Start:   time.Now(),
	}
	if err := database.SaveRecord(&record); err != nil {
		ProcessError(c, http.StatusInternalServerError, "failed to save record "+err.Error())
		return
	}
	models.TrackingActive(project)
	c.JSON(http.StatusOK, project)
}

func stop() error {
	records, err := database.GetAllRecords()
	if err != nil {
		return fmt.Errorf("failed to retrieve records %w", err)
	}
	for _, record := range records {
		if record.End.IsZero() {
			record.End = time.Now()
			if err := database.SaveRecord(&record); err != nil {
				log.Println("failed to save updated record", err)
			}
		}
	}
	models.TrackingInactive()
	return nil
}

func Stop(c *gin.Context) {
	if err := stop(); err != nil {
		ProcessError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, nil)
}

func Status(c *gin.Context) {
	status, err := GetStatus()
	if err != nil {
		ProcessError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, status)
}
