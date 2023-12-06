package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func getStatus(user string) (models.StatusResponse, error) {
	durations := make(map[string]time.Duration)
	status := models.Status{}
	response := models.StatusResponse{}
	records, err := database.GetTodaysRecordsForUser(user)
	if err != nil {
		return response, err
	}
	status.Current = models.Tracked(user)
	for _, record := range records {
		if record.End.IsZero() {
			record.End = time.Now()
			status.Elapsed = record.Duration()
		}
		durations[record.Project] = durations[record.Project] + record.End.Sub(record.Start)
		status.DailyTotal = status.DailyTotal + record.Duration()
		if record.Project == status.Current {
			status.Total = status.Total + record.Duration()
		}
	}
	response.Current = status.Current
	response.Elapsed = models.FmtDuration(status.Elapsed)
	response.CurrentTotal = models.FmtDuration(status.Total)
	response.DailyTotal = models.FmtDuration(status.DailyTotal)
	for k := range durations {
		value := models.FmtDuration(durations[k])
		duration := models.Duration{
			Project: k,
			Elapsed: value,
		}
		response.Durations = append(response.Durations, duration)
	}
	return response, nil
}

func getRecord(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		processError(c, 100, err.Error())
	}
	record, err := database.GetRecord(id)
	if err != nil {
		processError(c, 100, err.Error())
	}
	c.HTML(http.StatusOK, "editRecord", record)
}

func editRecord(c *gin.Context) {
	var err error
	edit := models.EditRecord{}
	if err := c.Bind(&edit); err != nil {
		processError(c, 100, err.Error())
		return
	}
	record, err := database.GetRecord(uuid.MustParse(edit.ID))
	if err != nil {
		processError(c, 200, err.Error())
		return
	}
	if err := c.Bind(&edit); err != nil {
		processError(c, 300, err.Error())
		return
	}
	record.End, err = time.Parse("2006-01-0215:04", edit.End+edit.EndTime)
	if err != nil {
		processError(c, 400, err.Error())
		return
	}
	record.Start, err = time.Parse("2006-01-0215:04", edit.Start+edit.StartTime)
	if err != nil {
		processError(c, 500, err.Error())
		return
	}
	if err := database.SaveRecord(&record); err != nil {
		processError(c, 600, err.Error())
		return
	}
	location := url.URL{Path: "/"}
	c.Redirect(http.StatusFound, location.RequestURI())
}
