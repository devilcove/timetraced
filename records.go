package main

import (
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
)

func GetStatus() (models.StatusResponse, error) {
	durations := make(map[string]time.Duration)
	status := models.Status{}
	response := models.StatusResponse{}
	records, err := database.GetAllRecords()
	if err != nil {
		return response, err
	}
	today := TruncateToStart(time.Now())
	for _, record := range records {
		if record.Start.Before(today) {
			continue
		}
		if record.End.IsZero() {
			status.Current = record.Project
			record.End = time.Now()
			status.Elapsed = record.Duration()
		}
		durations[record.Project] = durations[record.Project] + record.End.Sub(record.Start)
		status.Total = status.Total + record.Duration()
	}
	response.Current = status.Current
	response.Elapsed = models.FmtDuration(status.Elapsed)
	response.Total = models.FmtDuration(status.Total)
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

func TruncateToStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
func TruncateToEnd(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}
