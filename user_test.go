package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	database.InitializeDatabase()
	defer database.Close()
	os.Exit(m.Run())
}

func TestGoodLogin(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	data := struct {
		Username string
		Password string
	}{
		Username: "admin",
		Password: "password",
	}
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotNil(t, w.Result().Cookies())
}

func TestBadLogin(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	data := struct {
		Username string
		Password string
	}{
		Username: "admin",
		Password: "helloworld",
	}
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, []*http.Cookie{}, w.Result().Cookies())
}

func TestLogout(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/logout", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, []*http.Cookie{}, w.Result().Cookies())
}

func TestGetAllUsers(t *testing.T) {
	cookie := testLogin()
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users", nil)
	req.AddCookie(cookie)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	body, _ := io.ReadAll(w.Result().Body)
	users := []models.User{}
	json.Unmarshal(body, &users)
	for _, user := range users {
		assert.Equal(t, "", user.Password)
	}
}

func TestDeleteUser(t *testing.T)   {}
func TestEditUser(t *testing.T)     {}
func TestAddUser(t *testing.T)      {}
func TestHashPassword(t *testing.T) {}
func TestValidateUser(t *testing.T) {}

func testLogin() *http.Cookie {
	router := setupRouter()
	w := httptest.NewRecorder()
	data := struct {
		Username string
		Password string
	}{
		Username: "admin",
		Password: "password",
	}
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "time" {
			return cookie
		}
	}
	return nil
}
