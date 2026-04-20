package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
)

func getReport(w http.ResponseWriter, r *http.Request) {
	var err error
	projectsToQuery := []string{}
	user := getRequestUser(r)
	dbRequest := models.DatabaseReportRequest{
		User: user.Username,
	}
	reportRequest := models.ReportRequest{
		Start:   r.FormValue("start"),
		End:     r.FormValue("end"),
		Project: r.FormValue("project"),
	}
	slog.Info("getReport", "request", reportRequest)
	dbRequest.Start, err = time.Parse("2006-01-02", reportRequest.Start)
	if err != nil {
		processError(w, http.StatusBadRequest, err.Error())
		return
	}
	dbRequest.End, err = time.Parse("2006-01-02", reportRequest.End)
	if err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if reportRequest.Project == "" {
		allProjects, err := database.GetAllProjects()
		if err != nil {
			processError(w, http.StatusInternalServerError, err.Error())
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
			processError(w, http.StatusInternalServerError, err.Error())
			return
		}
		var total time.Duration
		for _, d := range data {
			recordTotal := d.End.Sub(d.Start)
			total += recordTotal
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
	render(w, "results", displayRecords)
}

func report(w http.ResponseWriter, r *http.Request) {
	user := getRequestUser(r)
	page := populatePage(user.Username)
	render(w, "report", page)
}
