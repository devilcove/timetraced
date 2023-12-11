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
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var (
	router *gin.Engine
	w      *httptest.ResponseRecorder
)

func TestMain(m *testing.M) {
	setLogging()
	os.Setenv("DB_FILE", "test.db")
	database.InitializeDatabase()
	defer database.Close()
	checkDefaultUser()
	router = setupRouter()
	w = httptest.NewRecorder()
	os.Exit(m.Run())
}

func TestDisplayLogin(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegister(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/register", nil)
	assert.Nil(t, err)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	t.Run("existing", func(t *testing.T) {
		err := createTestUser(models.User{Username: "tester", Password: "testing"})
		assert.Nil(t, err)
		w := httptest.NewRecorder()
		user := models.User{Username: "tester", Password: "testing"}
		payload, err := json.Marshal(&user)
		assert.Nil(t, err)
		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(payload))
		assert.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		assert.Contains(t, string(body), "user exists")
	})
	t.Run("new", func(t *testing.T) {
		w := httptest.NewRecorder()
		user := models.User{Username: "tester3", Password: ""}
		payload, err := json.Marshal(&user)
		assert.Nil(t, err)
		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(payload))
		assert.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, err := io.ReadAll(w.Result().Body)
		assert.Nil(t, err)
		assert.Contains(t, string(body), "password cannot be blank")
	})
}

func TestAdminLogin(t *testing.T) {
	data := struct {
		Username string
		Password string
	}{
		Username: "admin",
		Password: "password",
	}
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotNil(t, w.Result().Cookies())
}

func TestNonAdminLogin(t *testing.T) {
	deleteAllUsers()
	err := createTestUser(models.User{Username: "tester", Password: "testing", IsAdmin: false})
	assert.Nil(t, err)
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
	assert.Contains(t, string(response), "invalid user")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NotNil(t, w.Result().Cookies())
}

func TestBadLogin(t *testing.T) {
	t.Run("bad pass", func(t *testing.T) {
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
		response, err := io.ReadAll(w.Body)
		assert.Nil(t, err)
		assert.Contains(t, string(response), "invalid user")
	})
	t.Run("invalid data", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/login", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, []*http.Cookie{}, w.Result().Cookies())
	})
	t.Run("invalid user", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := models.User{Username: "nosuchuser", Password: "testing"}
		payload, _ := json.Marshal(data)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(payload))
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, []*http.Cookie{}, w.Result().Cookies())
		body, _ := io.ReadAll(w.Result().Body)
		assert.Contains(t, string(body), "invalid user")

	})
}

func TestLogout(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/logout", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []*http.Cookie{}, w.Result().Cookies())
}

func TestGetAllUsers(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
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
	})
}

func TestDeleteUser(t *testing.T) {
	deleteAllUsers()
	err := createTestUser(models.User{Username: "tester", Password: "testing", IsAdmin: false})
	assert.Nil(t, err)
	err = createTestUser(models.User{Username: "tester2", Password: "testing", IsAdmin: false})
	assert.Nil(t, err)
	t.Run("non-admin delete", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "tester", Password: "testing"})
		assert.NotNil(t, cookie)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/users/tester2", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		body, err := io.ReadAll(w.Body)
		assert.Nil(t, err)
		assert.Contains(t, string(body), "not authorized to delete this user")
	})
	t.Run("admin delete", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/users/tester", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
	t.Run("delete non-existent user", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/users/tester", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		response, err := io.ReadAll(w.Body)
		assert.Nil(t, err)
		assert.Contains(t, string(response), "user does not exist")
	})
}
func TestEditUser(t *testing.T) {
	deleteAllUsers()
	err := createTestUser(models.User{Username: "tester", Password: "testing", IsAdmin: false})
	assert.Nil(t, err)
	err = createTestUser(models.User{Username: "tester2", Password: "testing", IsAdmin: false})
	assert.Nil(t, err)

	t.Run("adminGetUser", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		req, err := http.NewRequest(http.MethodGet, "/users/tester", nil)
		assert.Nil(t, err)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetSelf", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "tester", Password: "testing"})
		assert.NotNil(t, cookie)
		req, err := http.NewRequest(http.MethodGet, "/users/tester", nil)
		assert.Nil(t, err)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetOther", func(t *testing.T) {
		w := httptest.NewRecorder()
		cookie := testLogin(models.User{Username: "tester", Password: "testing"})
		assert.NotNil(t, cookie)
		req, err := http.NewRequest(http.MethodGet, "/users/tester2", nil)
		assert.Nil(t, err)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		response, err := io.ReadAll(w.Body)
		assert.Nil(t, err)
		assert.Contains(t, string(response), "non-admin cannot edit other users")
	})

	t.Run("edit other user by non-admin", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "tester", Password: "testing"})
		assert.NotNil(t, cookie)
		w := httptest.NewRecorder()
		body, _ := json.Marshal(models.User{Username: "tester2", Password: "newPassword"})
		req, _ := http.NewRequest(http.MethodPost, "/users/tester2", bytes.NewBuffer(body))
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		response, err := io.ReadAll(w.Body)
		assert.Nil(t, err)
		assert.Contains(t, string(response), "not authorized to edit this user")
	})
	t.Run("edit user by admin", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)
		w := httptest.NewRecorder()
		body, _ := json.Marshal(models.User{Username: "tester2", Password: "newPassword"})
		req, _ := http.NewRequest(http.MethodPost, "/users/tester2", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		cookie = testLogin(models.User{Username: "tester2", Password: "newPassword"})
		assert.NotNil(t, cookie)
	})
	t.Run("edit user by self", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "tester", Password: "testing"})
		assert.NotNil(t, cookie)

		w := httptest.NewRecorder()
		body, _ := json.Marshal(models.User{Username: "tester", Password: "newPassword"})
		req, _ := http.NewRequest(http.MethodPost, "/users/tester", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		cookie = testLogin(models.User{Username: "tester", Password: "newPassword"})
		assert.NotNil(t, cookie)
	})
	t.Run("incomplete data", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/users/tester", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, _ := io.ReadAll(w.Result().Body)
		assert.Contains(t, string(body), "could not decode request into json")
	})
	t.Run("user does not exist", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)

		w := httptest.NewRecorder()
		payload, _ := json.Marshal(models.User{Username: "nosuchuser", Password: "newPassword"})
		req, _ := http.NewRequest(http.MethodPost, "/users/nosuchuser", bytes.NewBuffer(payload))
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, _ := io.ReadAll(w.Result().Body)
		assert.Contains(t, string(body), "user does not exist")
	})

}
func TestAddUser(t *testing.T) {
	deleteAllUsers()
	t.Run("add user by admin", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)

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
		err := createTestUser(models.User{Username: "tester", Password: "newPassword"})
		assert.Nil(t, err)
		cookie := testLogin(models.User{Username: "tester", Password: "newPassword"})
		assert.NotNil(t, cookie)

		w := httptest.NewRecorder()
		body, _ := json.Marshal(models.User{Username: "new", Password: "newPassword"})
		req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		body, _ = io.ReadAll(w.Body)
		assert.Contains(t, string(body), "only admins can create new users")
	})

	t.Run("empty password", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)

		w := httptest.NewRecorder()
		body, _ := json.Marshal(models.User{Username: "emptypass", Password: ""})
		req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, _ = io.ReadAll(w.Result().Body)
		assert.Contains(t, string(body), "username or password cannot be blank")
	})

	t.Run("incomplete data", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "admin", Password: "password"})
		assert.NotNil(t, cookie)

		w := httptest.NewRecorder()
		//body, _ := json.Marshal(struct{ InvaildData string }{InvaildData: "emptypass"})
		req, _ := http.NewRequest(http.MethodPost, "/users", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, _ := io.ReadAll(w.Result().Body)
		assert.Contains(t, string(body), "could not decode request into json")
	})

}

func testLogin(data models.User) *http.Cookie {

	w := httptest.NewRecorder()
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "time" {
			return cookie
		}
	}
	return nil
}

func createTestUser(user models.User) error {
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
