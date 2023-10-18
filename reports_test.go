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
	addTestUser(models.User{Username: "test", Password: "testing"})
	cookie := testLogin(models.User{Username: "test", Password: "testing"})
	assert.NotNil(t, cookie)
	t.Run("no request", func(t *testing.T) {
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/reports", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		body, _ := io.ReadAll(w.Result().Body)
		msg := models.ErrorMessage{}
		err := json.Unmarshal(body, &msg)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "could not decode request", msg.Message)
	})
	t.Run("no records", func(t *testing.T) {
		router := setupRouter()
		w := httptest.NewRecorder()
		data := models.ReportRequest{
			Start: time.Now().Add(-24 * time.Hour),
			End:   time.Now(),
		}
		payload, err := json.Marshal(data)
		assert.Nil(t, err)
		req, _ := http.NewRequest(http.MethodPost, "/reports", bytes.NewBuffer(payload))
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		body, _ := io.ReadAll(w.Result().Body)
		//msg := models.Report{}
		//err = json.Unmarshal(body, &msg)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		//t.Log(msg)
		//t.Log(string(body))
		assert.Equal(t, 2, len(body))
	})

	t.Run("one user/one project", func(t *testing.T) {
		createTestRecords()
		router := setupRouter()
		w := httptest.NewRecorder()
		data := models.ReportRequest{
			Start:    time.Now().Add(-24 * time.Hour),
			End:      time.Now(),
			Projects: []string{"timetrace"},
			Users:    []string{"test"},
		}
		payload, err := json.Marshal(data)
		assert.Nil(t, err)
		req, _ := http.NewRequest(http.MethodPost, "/reports", bytes.NewBuffer(payload))
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		body, _ := io.ReadAll(w.Result().Body)
		msg := []models.Report{}
		err = json.Unmarshal(body, &msg)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, len(msg))
	})

	t.Run("all records", func(t *testing.T) {
		router := setupRouter()
		w := httptest.NewRecorder()
		data := models.ReportRequest{
			Start:    time.Now().Add(-36 * time.Hour),
			End:      time.Now(),
			Projects: []string{"timetrace", "golf"},
			Users:    []string{"test", "test2"},
		}
		payload, err := json.Marshal(data)
		assert.Nil(t, err)
		req, _ := http.NewRequest(http.MethodPost, "/reports", bytes.NewBuffer(payload))
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		body, _ := io.ReadAll(w.Result().Body)
		msg := []models.Report{}
		err = json.Unmarshal(body, &msg)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 6, len(msg))
	})

}

func createTestRecords() {
	database.SaveRecord(&models.Record{
		ID:      uuid.New(),
		Project: "timetrace",
		User:    "test",
		Start:   time.Now().Add(time.Hour * -48),
		End:     time.Now().Add(time.Hour * -47),
	})
	database.SaveRecord(&models.Record{
		ID:      uuid.New(),
		Project: "timetrace",
		User:    "test",
		Start:   time.Now().Add(time.Hour * -49),
		End:     time.Now().Add(time.Hour * -48),
	})
	database.SaveRecord(&models.Record{
		ID:      uuid.New(),
		Project: "timetrace",
		User:    "test",
		Start:   time.Now().Add(time.Hour * -24),
		End:     time.Now().Add(time.Hour * -23),
	})
	database.SaveRecord(&models.Record{
		ID:      uuid.New(),
		Project: "golf",
		User:    "test",
		Start:   time.Now().Add(time.Hour * -48),
		End:     time.Now().Add(time.Hour * -47),
	})
	database.SaveRecord(&models.Record{
		ID:      uuid.New(),
		Project: "golf",
		User:    "test",
		Start:   time.Now().Add(time.Hour * -24),
		End:     time.Now().Add(time.Hour * -23),
	})
	database.SaveRecord(&models.Record{
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
		database.DeleteRecord(record.ID)
	}
}
