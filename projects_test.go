package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
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
		payload, err := json.Marshal(models.Project{
			Name: "test",
		})
		assert.Nil(t, err)
		req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(payload))
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		project := models.Project{}
		err = json.Unmarshal(body, &project)
		assert.Nil(t, err)
		assert.Equal(t, "test", project.Name)
		assert.Equal(t, true, regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(project.Name))
	})

	t.Run("invalid data", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/projects", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		msg := models.ErrorMessage{}
		err = json.Unmarshal(body, &msg)
		assert.Nil(t, err)
		assert.Equal(t, "could not decode request into json invalid request", msg.Message)
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
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		msg := models.ErrorMessage{}
		err = json.Unmarshal(body, &msg)
		assert.Nil(t, err)
		assert.Equal(t, "invalid project name", msg.Message)
	})

	t.Run("project exists", func(t *testing.T) {
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
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		msg := models.ErrorMessage{}
		err = json.Unmarshal(body, &msg)
		assert.Nil(t, err)
		assert.Equal(t, "project exists", msg.Message)
	})

}

func TestGetProjects(t *testing.T) {
	deleteAllProjects()
	createTestProjects()

	t.Run("existing project", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
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
		msg := models.ErrorMessage{}
		err = json.Unmarshal(body, &msg)
		assert.Nil(t, err)
		assert.Equal(t, "could not retrieve project unexpected end of JSON input", msg.Message)
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
		assert.Equal(t, "test", msg[0].Name)
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
	cookie := testLogin(models.User{Username: "admin", Password: "password"})
	assert.NotNil(t, cookie)
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/projects/status", nil)
	req.AddCookie(cookie)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	body, err := io.ReadAll(w.Result().Body)
	assert.Nil(t, err)
	msg := models.StatusResponse{}
	err = json.Unmarshal(body, &msg)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(msg.Durations))
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
		msg := models.ErrorMessage{}
		err = json.Unmarshal(body, &msg)
		assert.Nil(t, err)
		assert.Equal(t, "error reading project unexpected end of JSON input", msg.Message)

	})
	t.Run("inactive project", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/projects/inactive/start", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		msg := models.ErrorMessage{}
		err = json.Unmarshal(body, &msg)
		assert.Nil(t, err)
		assert.Equal(t, "project is not active", msg.Message)
	})
	t.Run("start", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/projects/test/start", nil)
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
	t.Run("start project already being tracked", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/projects/test2/start", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		msg := models.Project{}
		err = json.Unmarshal(body, &msg)
		assert.Nil(t, err)
		assert.Equal(t, "test2", msg.Name)
	})
	t.Run("stop tracked projects", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/projects/stop", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "null", string(body))
		t.Log(string(body))
	})
	t.Run("stop untracked project", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/projects/stop", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "null", string(body))
		t.Log(string(body))
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
}
