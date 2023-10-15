package main

import (
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
)

func getStatus() (models.StatusResponse, error) {
	durations := make(map[string]time.Duration)
	status := models.Status{}
	response := models.StatusResponse{}
	records, err := database.GetTodaysRecords()
	if err != nil {
		return response, err
	}
	status.Current = models.Tracked()
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
