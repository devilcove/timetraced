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

func TestAddProject(t *testing.T) {
	deleteAllProjects()

	t.Run("new project", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		project := models.Project{
			Name: "test",
		}
		payload, err := json.Marshal(&project)
		assert.Nil(t, err)
		t.Log(string(payload))
		req, err := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(payload))
		assert.Nil(t, err)
		req.AddCookie(cookie)
		req.Header.Set("Content-Type", "application/json")
		b, err := io.ReadAll(req.Body)
		assert.Nil(t, err)
		t.Log(string(b))
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		t.Log(string(body))
		assert.Contains(t, string(body), "")
	})

	t.Run("invalid data", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/projects", nil)
		req.AddCookie(cookie)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		assert.Contains(t, string(body), "could not decode request")
	})

	t.Run("invalid data2", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		payload, err := json.Marshal(models.Project{
			Name: "test name",
		})
		assert.Nil(t, err)
		req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(payload))
		req.AddCookie(cookie)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		assert.Contains(t, string(body), "invalid project name")
	})

	t.Run("project exists", func(t *testing.T) {
		createTestProjects()
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		payload, err := json.Marshal(models.Project{
			Name: "test",
		})
		assert.Nil(t, err)
		req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(payload))
		req.AddCookie(cookie)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		assert.Contains(t, string(body), "project exists")
	})

}

func TestGetProjects(t *testing.T) {
	deleteAllProjects()
	createTestProjects()
	createTestUser(models.User{Username: "test", Password: "test", IsAdmin: false})

	t.Run("existing project", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "test", Password: "test"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/projects/test", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		msg := models.Project{}
		err = json.Unmarshal(body, &msg)
		assert.Nil(t, err)
		assert.Equal(t, "test", msg.Name)
	})

	t.Run("wrong project", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/projects/missing", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		assert.Contains(t, string(body), "could not retrieve project unexpected end of JSON input")
	})

	t.Run("get all", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/projects", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		msg := []models.Project{}
		err = json.Unmarshal(body, &msg)
		assert.Nil(t, err)
		assert.Equal(t, 5, len(msg))
	})
	t.Run("get all when empty", func(t *testing.T) {
		deleteAllProjects()
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/projects", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		msg := []models.StatusResponse{}
		err = json.Unmarshal(body, &msg)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(msg))
	})
}

func TestGetStatus(t *testing.T) {
	createTestRecords()
	createTestUser(models.User{Username: "test", Password: "test", IsAdmin: false})
	cookie := testLogin(models.User{Username: "test", Password: "test"})
	assert.NotNil(t, cookie)
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/projects/status", nil)
	req.AddCookie(cookie)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	body, err := io.ReadAll(w.Result().Body)
	assert.Nil(t, err)
	assert.Contains(t, string(body), "<title>Time Tracking</title>")
}

func TestStartStopProject(t *testing.T) {
	deleteAllProjects()
	createTestProjects()
	t.Run("non-existent Project", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/projects/junk/start", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		assert.Contains(t, string(body), "error reading project unexpected end of JSON input")
	})
}

func deleteAllProjects() {
	projects, _ := database.GetAllProjects()
	for _, p := range projects {
		database.DeleteProject(p.Name)
	}
}

func createTestProjects() {
	database.SaveProject(&models.Project{
		ID:      uuid.New(),
		Name:    "test",
		Active:  true,
		Updated: time.Now(),
	})
	database.SaveProject(&models.Project{
		ID:      uuid.New(),
		Name:    "test2",
		Active:  true,
		Updated: time.Now(),
	})
	database.SaveProject(&models.Project{
		ID:      uuid.New(),
		Name:    "inactive",
		Active:  false,
		Updated: time.Now(),
	})
	database.SaveProject(&models.Project{
		ID:      uuid.New(),
		Name:    "timetrace",
		Active:  false,
		Updated: time.Now(),
	})
	database.SaveProject(&models.Project{
		ID:      uuid.New(),
		Name:    "golf",
		Active:  false,
		Updated: time.Now(),
	})
}
