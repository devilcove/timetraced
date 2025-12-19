package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kairum-Labs/should"
	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/google/uuid"
)

func TestAddProject(t *testing.T) {
	deleteAllProjects()
	createAdmin()

	t.Run("displayForm", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/projects/add/", nil)
		r.AddCookie(adminLogin())
		router.ServeHTTP(w, r)
		should.BeEqual(t, w.Result().StatusCode, http.StatusOK)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "Add New Project")
	})
	t.Run("new", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := bodyParams("name", "test")
		req := httptest.NewRequest(http.MethodPost, "/projects/", payload)
		req.AddCookie(adminLogin())
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusOK)
		projects, err := database.GetAllProjects()
		should.BeNil(t, err)
		should.BeEqual(t, len(projects), 1)
	})

	t.Run("empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/projects/", nil)
		req.AddCookie(adminLogin())
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "invalid project name")
	})

	t.Run("invalidData", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload, err := json.Marshal(models.Project{
			Name: "test name",
		})
		should.BeNil(t, err)
		req, _ := http.NewRequest(http.MethodPost, "/projects/", bytes.NewBuffer(payload))
		req.AddCookie(adminLogin())
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "invalid project name")
	})

	t.Run("exists", func(t *testing.T) {
		createTestProjects()
		w := httptest.NewRecorder()
		payload := bodyParams("name", "test")
		req := httptest.NewRequest(http.MethodPost, "/projects/", payload)
		req.AddCookie(adminLogin())
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "project exists")
	})
}

func TestStartStopProject(t *testing.T) {
	deleteAllRecords()
	deleteAllProjects()
	createTestProjects()
	deleteAllUsers()
	createAdmin()
	t.Run("non-existent", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/projects/start/junk", nil)
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "no such project")
	})
	t.Run("inactive", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/projects/start/inactive", nil)
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "project is not active")
	})
	t.Run("active", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/projects/start/test", nil)
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusOK)
		records, err := database.GetAllRecords()
		should.BeNil(t, err)
		should.BeEqual(t, len(records), 1)
		should.BeEqual(t, records[0].End.IsZero(), true)
	})
	t.Run("startdifferent", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/projects/start/test2", nil)
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusOK)
		records, err := database.GetAllRecords()
		should.BeNil(t, err)
		should.BeEqual(t, len(records), 2)
		for _, record := range records {
			if record.Project == "test" {
				should.BeEqual(t, record.End.IsZero(), false)
			}
			if record.Project == "test2" {
				should.BeEqual(t, record.End.IsZero(), true)
			}
		}
	})
	t.Run("stop", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/projects/stop/", nil)
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusOK)
		records, err := database.GetAllRecords()
		should.BeNil(t, err)
		should.BeEqual(t, len(records), 2)
		should.BeEqual(t, records[0].End.IsZero(), false)
		should.BeEqual(t, records[1].End.IsZero(), false)
	})
}

func deleteAllProjects() {
	projects, _ := database.GetAllProjects()
	for _, p := range projects {
		_ = database.DeleteProject(p.Name)
	}
}

func createTestProjects() {
	_ = database.SaveProject(&models.Project{
		ID:      uuid.New(),
		Name:    "test",
		Active:  true,
		Updated: time.Now(),
	})
	_ = database.SaveProject(&models.Project{
		ID:      uuid.New(),
		Name:    "test2",
		Active:  true,
		Updated: time.Now(),
	})
	_ = database.SaveProject(&models.Project{
		ID:      uuid.New(),
		Name:    "inactive",
		Active:  false,
		Updated: time.Now(),
	})
	_ = database.SaveProject(&models.Project{
		ID:      uuid.New(),
		Name:    "timetrace",
		Active:  false,
		Updated: time.Now(),
	})
	_ = database.SaveProject(&models.Project{
		ID:      uuid.New(),
		Name:    "golf",
		Active:  false,
		Updated: time.Now(),
	})
}
