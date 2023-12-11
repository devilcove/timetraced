package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetReport(t *testing.T) {
	deleteAllUsers()
	deleteAllRecords()
	err := createTestUser(models.User{Username: "test", Password: "testing"})
	assert.Nil(t, err)
	cookie := testLogin(models.User{Username: "test", Password: "testing"})
	assert.NotNil(t, cookie)
	t.Run("no request", func(t *testing.T) {

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/reports", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, string(body), "could not decode request")
	})
	t.Run("no records", func(t *testing.T) {

		w := httptest.NewRecorder()
		request := models.ReportRequest{
			Start:   time.Now().Add(-24 * time.Hour).Format("2006-01-02"),
			End:     time.Now().Format("2006-01-02"),
			Project: "nilProject",
		}
		payload, err := json.Marshal(&request)
		assert.Nil(t, err)
		req, _ := http.NewRequest(http.MethodPost, "/reports", bytes.NewBuffer(payload))
		req.AddCookie(cookie)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotContains(t, string(body), "<tr>")
	})

	t.Run("one user/one project", func(t *testing.T) {
		createTestRecords()

		w := httptest.NewRecorder()
		data := models.ReportRequest{
			Start:   time.Now().Add(-24 * time.Hour).Format("2006-01-02"),
			End:     time.Now().Format("2006-01-02"),
			Project: "timetrace",
		}
		payload, err := json.Marshal(data)
		assert.Nil(t, err)
		req, _ := http.NewRequest(http.MethodPost, "/reports", bytes.NewBuffer(payload))
		req.AddCookie(cookie)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		body, _ := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, string(body), "<h2>Project timetrace")
	})

	t.Run("all records", func(t *testing.T) {
		createTestRecords()
		createTestProjects()

		w := httptest.NewRecorder()
		data := models.ReportRequest{
			Start:   time.Now().Add(-72 * time.Hour).Format("2006-01-02"),
			End:     time.Now().Format("2006-01-02"),
			Project: "",
		}
		payload, err := json.Marshal(data)
		assert.Nil(t, err)
		req, _ := http.NewRequest(http.MethodPost, "/reports", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, string(body), "<button hx-get=\"/records")
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
