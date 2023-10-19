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
	createTestProject()
	t.Run("existing project", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/projects/test project", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		msg := models.Project{}
		err = json.Unmarshal(body, &msg)
		assert.Nil(t, err)
		assert.Equal(t, "test project", msg.Name)

	})
	t.Run("wrong project", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/projects/missing project", nil)
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

func deleteAllProjects() {
	projects, _ := database.GetAllProjects()
	for _, p := range projects {
		database.DeleteProject(p.Name)
	}
}

func createTestProject() {
	database.SaveProject(&models.Project{
		ID:      uuid.New(),
		Name:    "test project",
		Active:  true,
		Updated: time.Now(),
	})
}
