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
	setLogging()
	os.Setenv("DB_FILE", "test.db")
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
	assert.Equal(t, true, user.IsAdmin)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotNil(t, w.Result().Cookies())
}
func TestNonAdminLogin(t *testing.T) {
	deleteAllUsers()
	err := addTestUser(models.User{Username: "tester", Password: "testing", IsAdmin: false})
	assert.Nil(t, err)
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
	deleteAllUsers()
	err := addTestUser(models.User{Username: "tester", Password: "testing", IsAdmin: false})
	assert.Nil(t, err)
	err = addTestUser(models.User{Username: "tester2", Password: "testing", IsAdmin: false})
	assert.Nil(t, err)
	t.Run("non-admin delete", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "tester", Password: "testing"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/users/tester2", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("admin delete", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/users/tester", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
	t.Run("delete non-existent user", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/users/tester", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
func TestEditUser(t *testing.T) {
	deleteAllUsers()
	err := addTestUser(models.User{Username: "tester", Password: "testing", IsAdmin: false})
	assert.Nil(t, err)
	err = addTestUser(models.User{Username: "tester2", Password: "testing", IsAdmin: false})
	assert.Nil(t, err)
	t.Run("edit other user by non-admin", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "tester", Password: "testing"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		body, _ := json.Marshal(models.User{Username: "tester2", Password: "newPassword"})
		req, _ := http.NewRequest(http.MethodPut, "/users", bytes.NewBuffer(body))
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("edit user by admin", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		body, _ := json.Marshal(models.User{Username: "tester2", Password: "newPassword"})
		req, _ := http.NewRequest(http.MethodPut, "/users", bytes.NewBuffer(body))
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		cookie = testLogin(models.User{Username: "tester2", Password: "newPassword"})
		assert.NotNil(t, cookie)
	})
	t.Run("edit user by self", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "tester", Password: "testing"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		body, _ := json.Marshal(models.User{Username: "tester", Password: "newPassword"})
		req, _ := http.NewRequest(http.MethodPut, "/users", bytes.NewBuffer(body))
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		cookie = testLogin(models.User{Username: "tester", Password: "newPassword"})
		assert.NotNil(t, cookie)
	})
}
func TestAddUser(t *testing.T) {
	deleteAllUsers()
	t.Run("add user by admin", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		body, _ := json.Marshal(models.User{Username: "new", Password: "newPassword"})
		req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		if assert.Equal(t, http.StatusNoContent, w.Code) {
			cookie = testLogin(models.User{Username: "new", Password: "newPassword"})
			assert.NotNil(t, cookie)
		} else {
			body, _ := io.ReadAll(w.Result().Body)
			t.Log(w.Code, string(body))
		}
	})
	t.Run("add user by non-admin", func(t *testing.T) {
		err := addTestUser(models.User{Username: "tester", Password: "newPassword"})
		assert.Nil(t, err)
		cookie := testLogin(models.User{Username: "tester", Password: "newPassword"})
		assert.NotNil(t, cookie)
		router := setupRouter()
		w := httptest.NewRecorder()
		body, _ := json.Marshal(models.User{Username: "new", Password: "newPassword"})
		req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

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

func addTestUser(user models.User) error {
	user.Password, _ = hashPassword(user.Password)
	if err := database.SaveUser(&user); err != nil {
		return err
	}
	return nil
}

func deleteAllUsers() {
	users, _ := database.GetAllUsers()
	for _, user := range users {
		if user.Username != "admin" {
			database.DeleteUser(user.Username)
		}
	}
}
