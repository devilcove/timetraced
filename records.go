package main

import (
	"net/http"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
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
		durations[record.Project] += record.End.Sub(record.Start)
		status.DailyTotal += record.Duration()
		if record.Project == status.Current {
			status.Total += record.Duration()
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

func getRecord(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		processError(w, http.StatusBadRequest, err.Error())
		return
	}
	record, err := database.GetRecord(id)
	if err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	templates.ExecuteTemplate(w, "editRecord", record)
}

func editRecord(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		processError(w, http.StatusBadRequest, "invalid data")
	}

	edit := models.EditRecord{
		ID:        r.PathValue("id"),
		Start:     r.FormValue("Start"),
		StartTime: r.FormValue("StartTime"),
		End:       r.FormValue("End"),
		EndTime:   r.FormValue("EndTime"),
	}
	record, err := database.GetRecord(uuid.MustParse(edit.ID))
	if err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	record.End, err = time.Parse("2006-01-0215:04", edit.End+edit.EndTime)
	if err != nil {
		processError(w, http.StatusBadRequest, err.Error())
		return
	}
	record.Start, err = time.Parse("2006-01-0215:04", edit.Start+edit.StartTime)
	if err != nil {
		processError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := database.SaveRecord(&record); err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	displayStatus(w, r)
}
