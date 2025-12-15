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
	"github.com/devilcove/timetraced/models"
	"github.com/google/uuid"
)

func TestGetReport(t *testing.T) {
	deleteAllUsers()
	deleteAllRecords()
	createAdmin()
	err := createTestUser(models.User{Username: "test", Password: "testing"})
	should.BeNil(t, err)
	cookie := testLogin(models.User{Username: "test", Password: "testing"})
	should.NotBeNil(t, cookie)

	t.Run("get", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/reports/", nil)
		r.AddCookie(adminLogin())
		router.ServeHTTP(w, r)
		should.BeEqual(t, w.Result().StatusCode, http.StatusOK)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "Reports")
	})
	t.Run("no request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/reports/", nil)
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		should.ContainSubstring(t, string(body), "parsing time")
	})
	t.Run("no records", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := bodyParams(
			"start", time.Now().Add(-24*time.Hour).Format("2006-01-02"),
			"end", time.Now().Format("2006-01-02"),
			"project", "nilProject",
		)
		req := httptest.NewRequest(http.MethodPost, "/reports/", payload)
		req.AddCookie(adminLogin())
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.BeEqual(t, w.Code, http.StatusOK)
		// should.NotContainSubstring(t, string(body), "<tr>")
		if strings.Contains(string(body), "<tr>") {
			t.Fail()
		}
	})

	t.Run("nilproject", func(t *testing.T) {
		createTestRecords()
		w := httptest.NewRecorder()
		payload := bodyParams(
			"start", time.Now().Add(-24*time.Hour).Format("2006-01-02"),
			"end", time.Now().Format("2006-01-02"),
			"project", "nilProject",
		)
		req := httptest.NewRequest(http.MethodPost, "/reports/", payload)
		req.AddCookie(cookie)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		body, _ := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.BeEqual(t, w.Code, http.StatusOK)
		should.ContainSubstring(t, string(body), "<h1>TimeTrace Report")
	})

	t.Run("allRecords", func(t *testing.T) {
		createTestRecords()
		createTestProjects()

		w := httptest.NewRecorder()
		payload := bodyParams(
			"start", time.Now().Add(-24*time.Hour*14).Format("2006-01-02"),
			"end", time.Now().Format("2006-01-02"),
		)
		req := httptest.NewRequest(http.MethodPost, "/reports/", payload)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.BeEqual(t, w.Code, http.StatusOK)
		should.ContainSubstring(t, string(body), "TimeTrace Report")
	})
}

func createTestRecords() {
	_ = database.SaveRecord(&models.Record{
		ID:      uuid.New(),
		Project: "timetrace",
		User:    "test",
		Start:   time.Now().Add(time.Minute * -10),
		End:     time.Now().Add(time.Minute * -5),
	})

	_ = database.SaveRecord(&models.Record{
		ID:      uuid.New(),
		Project: "timetrace",
		User:    "test",
		Start:   time.Now().Add(time.Hour * -48),
		End:     time.Now().Add(time.Hour * -47),
	})
	_ = database.SaveRecord(&models.Record{
		ID:      uuid.New(),
		Project: "timetrace",
		User:    "test",
		Start:   time.Now().Add(time.Hour * -49),
		End:     time.Now().Add(time.Hour * -48),
	})
	_ = database.SaveRecord(&models.Record{
		ID:      uuid.New(),
		Project: "timetrace",
		User:    "test",
		Start:   time.Now().Add(time.Hour * -24),
		End:     time.Now().Add(time.Hour * -23),
	})
	_ = database.SaveRecord(&models.Record{
		ID:      uuid.New(),
		Project: "golf",
		User:    "test",
		Start:   time.Now().Add(time.Hour * -48),
		End:     time.Now().Add(time.Hour * -47),
	})
	_ = database.SaveRecord(&models.Record{
		ID:      uuid.New(),
		Project: "golf",
		User:    "test",
		Start:   time.Now().Add(time.Hour * -24),
		End:     time.Now().Add(time.Hour * -23),
	})
	_ = database.SaveRecord(&models.Record{
		ID:      uuid.New(),
		Project: "timetrace",
		User:    "test2",
		Start:   time.Now().Add(time.Hour * -48),
		End:     time.Now().Add(time.Hour * -47),
	})
}

func deleteAllRecords() {
	records, _ := database.GetAllRecords()
	for _, record := range records {
		_ = database.DeleteRecord(record.ID)
	}
}
