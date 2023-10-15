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
	checkDefaultUser()
	os.Exit(m.Run())
}

func TestAdminLogin(t *testing.T) {
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
	response, err := io.ReadAll(w.Result().Body)
	assert.Nil(t, err)
	user := models.User{}
	err = json.Unmarshal(response, &user)
	assert.Nil(t, err)
	assert.Nil(t, err)
	assert.Equal(t, true, user.IsAdmin)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotNil(t, w.Result().Cookies())
}
func TestNonAdminLogin(t *testing.T) {
	cookie := testLogin(models.User{Username: "admin", Password: "password"})
	addTestUser(cookie, models.User{Username: "tester", Password: "testing", IsAdmin: false})
	router := setupRouter()
	w := httptest.NewRecorder()
	data := models.User{
		Username: "tester",
		Password: "testing",
	}
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)
	response, err := io.ReadAll(w.Result().Body)
	assert.Nil(t, err)
	user := models.User{}
	err = json.Unmarshal(response, &user)
	assert.Nil(t, err)
	assert.Nil(t, err)
	assert.Equal(t, false, user.IsAdmin)
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
	cookie := testLogin(models.User{Username: "admin", Password: "password"})
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

func TestDeleteUser(t *testing.T) {
	t.Log("login as admin")
	cookie := testLogin(models.User{Username: "admin", Password: "password"})
	t.Log("create tester")
	addTestUser(cookie, models.User{Username: "tester", Password: "testing", IsAdmin: false})
	t.Log("create tester2")
	addTestUser(cookie, models.User{Username: "tester2", Password: "testing", IsAdmin: false})
	t.Run("non-admin delete", func(t *testing.T) {
		t.Log("non admin delete")
		t.Log("login as tester")
		cookie = testLogin(models.User{Username: "tester", Password: "testing"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/users/tester2", nil)
		req.AddCookie(cookie)
		t.Log("attempt to delete tester2")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("admin delete", func(t *testing.T) {
		t.Log("admin delete")
		t.Log("log in as admin")
		cookie = testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/users/tester", nil)
		req.AddCookie(cookie)
		t.Log("delete tester")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
		t.Log(t, w.Result().Body)
	})
	t.Run("delete non-existent user", func(t *testing.T) {
		t.Log("delete non-existent user")
		cookie = testLogin(models.User{Username: "admin", Password: "password"})
		t.Log("log in as admin")
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/users/tester", nil)
		req.AddCookie(cookie)
		t.Log("delete tester")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
func TestEditUser(t *testing.T)     {}
func TestAddUser(t *testing.T)      {}
func TestHashPassword(t *testing.T) {}
func TestValidateUser(t *testing.T) {}

func testLogin(data models.User) *http.Cookie {
	router := setupRouter()
	w := httptest.NewRecorder()
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

func addTestUser(cookie *http.Cookie, user models.User) {
	router := setupRouter()
	w := httptest.NewRecorder()
	body, _ := json.Marshal(user)
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.AddCookie(cookie)
	router.ServeHTTP(w, req)
}
