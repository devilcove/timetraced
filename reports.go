package main

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func getReport(c *gin.Context) {
	var err error
	projectsToQuery := []string{}
	session := sessions.Default(c)
	dbRequest := models.DatabaseReportRequest{
		User: session.Get("user").(string),
	}
	reportRequest := models.ReportRequest{}
	if err := c.Bind(&reportRequest); err != nil {
		slog.Error("unable to bind", "error", err)
		processError(c, "bad request", "could not decode request")
		if c.Request.Body != nil {
			body, _ := io.ReadAll(c.Request.Body)
			slog.Info(string(body))
		}
		return
	}

	dbRequest.Start, err = time.Parse("2006-01-02", reportRequest.Start)
	if err != nil {
		processError(c, "ServerError", err.Error())
		return
	}
	dbRequest.End, err = time.Parse("2006-01-02", reportRequest.End)
	if err != nil {
		processError(c, "ServerError", err.Error())
		return
	}
	if reportRequest.Project == "" {
		allProjects, err := database.GetAllProjects()
		if err != nil {
			processError(c, "ServerError", err.Error())
			return
		}
		for _, project := range allProjects {
			projectsToQuery = append(projectsToQuery, project.Name)
		}
	} else {
		projectsToQuery = append(projectsToQuery, reportRequest.Project)
	}
	displayRecords := []models.Report{}
	for _, project := range projectsToQuery {
		displayRecord := models.Report{}
		reportRecord := models.ReportRecord{}
		reportRecords := []models.ReportRecord{}
		dbRequest.Project = project
		data, err := database.GetReportRecords(dbRequest)
		if err != nil {
			processError(c, "ServerError", err.Error())
			return
		}
		var total time.Duration
		for _, d := range data {
			total = d.End.Sub(d.Start)
			reportRecord.End = d.End
			reportRecord.Start = d.Start
			reportRecord.ID = d.ID
			reportRecords = append(reportRecords, reportRecord)
		}
		if total != 0 {
			displayRecord.Project = project
			displayRecord.Total = models.FmtDuration(total)
			displayRecord.Items = reportRecords
			displayRecords = append(displayRecords, displayRecord)
		}
	}
	c.HTML(http.StatusOK, "results", displayRecords)

}

func report(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user").(string)
	page := populatePage(user)
	projects, err := database.GetAllProjects()
	if err != nil {
		slog.Error(err.Error())
	} else {
		for _, project := range projects {
			page.Projects = append(page.Projects, project.Name)
		}
	}
	c.HTML(http.StatusOK, "report", page)
}
