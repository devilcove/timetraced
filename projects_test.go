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
	err := createTestUser(models.User{Username: "admin", Password: "password", IsAdmin: true})
	should.BeNil(t, err)

	t.Run("new project", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		should.NotBeNil(t, cookie)

		w := httptest.NewRecorder()
		project := models.Project{
			Name: "test",
		}
		payload, err := json.Marshal(&project)
		should.BeNil(t, err)
		t.Log(string(payload))
		req, err := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(payload))
		should.BeNil(t, err)
		req.AddCookie(cookie)
		req.Header.Set("Content-Type", "application/json")
		b, err := io.ReadAll(req.Body)
		should.BeNil(t, err)
		t.Log(string(b))
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		t.Log(string(body))
		should.ContainSubstring(t, string(body), "")
	})

	t.Run("invalid data", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		should.NotBeNil(t, cookie)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/projects", nil)
		req.AddCookie(cookie)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "could not decode request")
	})

	t.Run("invalid data2", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		should.NotBeNil(t, cookie)

		w := httptest.NewRecorder()
		payload, err := json.Marshal(models.Project{
			Name: "test name",
		})
		should.BeNil(t, err)
		req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(payload))
		req.AddCookie(cookie)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "invalid project name")
	})

	t.Run("project exists", func(t *testing.T) {
		createTestProjects()
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		should.NotBeNil(t, cookie)

		w := httptest.NewRecorder()
		payload, err := json.Marshal(models.Project{
			Name: "test",
		})
		should.BeNil(t, err)
		req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(payload))
		req.AddCookie(cookie)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "project exists")
	})
}

func TestGetProjects(t *testing.T) {
	deleteAllProjects()
	createTestProjects()
	err := createTestUser(models.User{Username: "test", Password: "test", IsAdmin: false})
	should.BeNil(t, err)

	t.Run("existing project", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "test", Password: "test"})
		should.NotBeNil(t, cookie)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/projects/test", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusOK)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		msg := models.Project{}
		err = json.Unmarshal(body, &msg)
		should.BeNil(t, err)
		should.BeEqual(t, msg.Name, "test")
	})

	t.Run("wrong project", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		should.NotBeNil(t, cookie)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/projects/missing", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "could not retrieve project no such project")
	})

	t.Run("get all", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		should.NotBeNil(t, cookie)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/projects", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusOK)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		msg := []models.Project{}
		err = json.Unmarshal(body, &msg)
		should.BeNil(t, err)
		should.BeEqual(t, len(msg), 5)
	})
	t.Run("get all when empty", func(t *testing.T) {
		deleteAllProjects()
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		should.NotBeNil(t, cookie)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/projects", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusOK)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		msg := []models.StatusResponse{}
		err = json.Unmarshal(body, &msg)
		should.BeNil(t, err)
		should.BeEqual(t, len(msg), 0)
	})
}

func TestGetStatus(t *testing.T) {
	createTestRecords()
	err := createTestUser(models.User{Username: "test", Password: "test", IsAdmin: false})
	should.BeNil(t, err)
	cookie := testLogin(models.User{Username: "test", Password: "test"})
	should.NotBeNil(t, cookie)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/projects/status", nil)
	req.AddCookie(cookie)
	router.ServeHTTP(w, req)
	should.BeEqual(t, w.Code, http.StatusOK)
	body, err := io.ReadAll(w.Result().Body)
	should.BeNil(t, err)
	should.ContainSubstring(t, string(body), "<b>Current Project: </b>")
}

func TestStartStopProject(t *testing.T) {
	deleteAllProjects()
	createTestProjects()
	t.Run("non-existent Project", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		should.NotBeNil(t, cookie)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/projects/junk/start", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusInternalServerError)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "no such project")
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
