package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Kairum-Labs/should"
	"github.com/devilcove/timetraced/database"
)

func TestRecords(t *testing.T) {
	deleteAllRecords()
	createTestRecords()
	deleteAllUsers()
	createAdmin()
	records, err := database.GetAllRecords()
	should.BeNil(t, err)
	ID := records[0].ID.String()
	url := "/records/" + ID
	t.Run("get", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, url, nil)
		r.AddCookie(adminLogin())
		router.ServeHTTP(w, r)
		should.BeEqual(t, w.Result().StatusCode, http.StatusOK)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "Edit Record")
	})
	t.Run("invalidID", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/records/notUUID", nil)
		r.AddCookie(adminLogin())
		router.ServeHTTP(w, r)
		should.BeEqual(t, w.Result().StatusCode, http.StatusBadRequest)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "invalid UUID")
	})
	t.Run("edit", func(t *testing.T) {
		start := time.Now().Add(time.Hour - 1)
		end := time.Now()
		w := httptest.NewRecorder()
		payload := bodyParams(
			"ID", ID,
			"Start", start.Format(time.DateOnly),
			"StartTime", formatTimeOnly(start),
			"End", end.Format(time.DateOnly),
			"EndTime", formatTimeOnly(end),
		)
		r := httptest.NewRequest(http.MethodPost, url, payload)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.AddCookie(adminLogin())
		router.ServeHTTP(w, r)
		should.BeEqual(t, w.Result().StatusCode, http.StatusOK)
		record, err := database.GetRecord(records[0].ID)
		t.Log("record", record.Start, "start", start)
		should.BeNil(t, err)
		should.BeEqual(t, record.Start.Format(time.DateOnly), start.Format(time.DateOnly))
		should.BeEqual(t, record.End.Format(time.DateOnly), end.Format(time.DateOnly))
	})
	t.Run("editBadID", func(t *testing.T) {
		start := time.Now().Add(time.Hour - 1)
		end := time.Now()
		w := httptest.NewRecorder()
		payload := bodyParams(
			"ID", "notUUID",
			"Start", start.Format(time.DateOnly),
			"StartTime", formatTimeOnly(start),
			"End", end.Format(time.DateOnly),
			"EndTime", formatTimeOnly(end),
		)
		r := httptest.NewRequest(http.MethodPost, "/records/notUUID", payload)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.AddCookie(adminLogin())
		should.Panic(t, func() {
			router.ServeHTTP(w, r)
		})
		should.BeEqual(t, w.Result().StatusCode, http.StatusOK)
	})

	t.Run("badStartTime", func(t *testing.T) {
		start := time.Now().Add(time.Hour - 1)
		end := time.Now()
		w := httptest.NewRecorder()
		payload := bodyParams(
			"ID", ID,
			"Start", start.Format(time.DateOnly),
			"StartTime", start.Format(time.TimeOnly),
			"End", end.Format(time.DateOnly),
			"EndTime", formatTimeOnly(end),
		)
		r := httptest.NewRequest(http.MethodPost, url, payload)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.AddCookie(adminLogin())
		router.ServeHTTP(w, r)
		should.BeEqual(t, w.Result().StatusCode, http.StatusBadRequest)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "parsing time")
	})

	t.Run("badEndTime", func(t *testing.T) {
		start := time.Now().Add(time.Hour - 1)
		end := time.Now()
		w := httptest.NewRecorder()
		payload := bodyParams(
			"ID", ID,
			"Start", start.Format(time.DateOnly),
			"StartTime", formatTimeOnly(start),
			"End", end.Format(time.DateOnly),
			"EndTime", end.Format(time.TimeOnly),
		)
		r := httptest.NewRequest(http.MethodPost, url, payload)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.AddCookie(adminLogin())
		router.ServeHTTP(w, r)
		should.BeEqual(t, w.Result().StatusCode, http.StatusBadRequest)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "parsing time")
	})
}

func formatTimeOnly(t time.Time) string {
	s := t.Format(time.TimeOnly)
	index := strings.LastIndex(s, ":")
	return s[:index]
}
