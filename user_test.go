package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Kairum-Labs/should"
	"github.com/devilcove/mux"
	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/mattkasun/tools/logging"
)

var (
	router *mux.Router
	w      *httptest.ResponseRecorder
)

func TestMain(m *testing.M) {
	log := logging.TextLogger(logging.TruncateSource(), logging.TimeFormat(time.DateTime))
	os.Setenv("USER", "")
	os.Setenv("PASS", "")
	os.Setenv("DB_FILE", "test.db") //nolint:errcheck,gosec
	_ = database.InitializeDatabase()
	defer database.Close()
	router = setupRouter(log.Logger)
	w = httptest.NewRecorder()
	os.Exit(m.Run())
}

func TestDisplayLogin(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	should.BeNil(t, err)
	router.ServeHTTP(w, req)
	should.BeEqual(t, w.Code, http.StatusOK)
}

func TestGetAllUsers(t *testing.T) {
	deleteAllUsers()
	createAdmin()
	createTestUser(models.User{Username: "test", Password: "pass"})
	t.Run("admin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/", nil)
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusOK)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "<h1>Users</h1>")
	})
	t.Run("normalUser", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/users/", nil)
		r.AddCookie(testLogin(models.User{Username: "test", Password: "pass"}))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		should.BeEqual(t, w.Result().StatusCode, http.StatusOK)
		body, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "function valPass")
	})
}

func TestRegister(t *testing.T) {
	deleteAllUsers()
	createAdmin()
	req := httptest.NewRequest(http.MethodGet, "/users/register/", nil)
	req.AddCookie(adminLogin())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	should.BeEqual(t, w.Code, http.StatusOK)
	t.Run("existing", func(t *testing.T) {
		err := createTestUser(models.User{Username: "tester", Password: "testing"})
		should.BeNil(t, err)
		w := httptest.NewRecorder()
		body := bodyParams("username", "tester", "password", "testing")
		req := httptest.NewRequest(http.MethodPost, "/users/register/", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		b, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(b), "user exists")
	})
	t.Run("blankPass", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bodyParams("username", "tester3", "password", "")
		req := httptest.NewRequest(http.MethodPost, "/users/register/", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		b, err := io.ReadAll(w.Result().Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(b), "password cannot be blank")
	})
	t.Run("new", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bodyParams("username", "tester3", "password", "pass")
		req := httptest.NewRequest(http.MethodPost, "/users/register/", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusFound)
		users, err := database.GetAllUsers()
		should.BeNil(t, err)
		should.BeEqual(t, len(users), 3)
	})
}

func TestAdminLogin(t *testing.T) {
	body := bodyParams("username", "admin", "password", "password")
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	should.BeEqual(t, w.Code, http.StatusOK)
	should.NotBeNil(t, w.Result().Cookies())
}

func TestNonAdminLogin(t *testing.T) {
	deleteAllUsers()
	err := createTestUser(models.User{Username: "tester", Password: "testing", IsAdmin: false})
	should.BeNil(t, err)
	w := httptest.NewRecorder()
	data := models.User{
		Username: "tester",
		Password: "testing",
	}
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)
	response, err := io.ReadAll(w.Result().Body)
	should.BeNil(t, err)
	should.ContainSubstring(t, string(response), "invalid user")
	should.BeEqual(t, w.Code, http.StatusBadRequest)
	should.NotBeNil(t, w.Result().Cookies())
}

func TestBadLogin(t *testing.T) {
	t.Run("bad pass", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := bodyParams("username", "admin", "password", "wrong")
		req := httptest.NewRequest(http.MethodPost, "/login", payload)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		should.BeEqual(t, w.Result().Cookies(), []*http.Cookie{})
		response, err := io.ReadAll(w.Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(response), "invalid user")
	})
	t.Run("invalid data", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		should.BeEqual(t, w.Result().Cookies(), []*http.Cookie{})
	})
	t.Run("invalid user", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := bodyParams("username", "nosuchuser", "passowd", "testing")
		req := httptest.NewRequest(http.MethodPost, "/login", payload)
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		should.BeEqual(t, w.Result().Cookies(), []*http.Cookie{})
		body, _ := io.ReadAll(w.Result().Body)
		should.ContainSubstring(t, string(body), "invalid user")
	})
}

func TestLogout(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/logout/", nil)
	router.ServeHTTP(w, req)
	should.BeEqual(t, w.Code, http.StatusFound)
	should.BeEqual(t, w.Result().Cookies()[0].MaxAge, -1)
}

func TestDeleteUser(t *testing.T) {
	deleteAllUsers()
	createAdmin()
	err := createTestUser(models.User{Username: "tester", Password: "testing", IsAdmin: false})
	should.BeNil(t, err)
	err = createTestUser(models.User{Username: "tester2", Password: "testing", IsAdmin: false})
	should.BeNil(t, err)
	t.Run("non-admin delete", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "tester", Password: "testing"})
		should.NotBeNil(t, cookie)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/users/tester2", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusUnauthorized)
		body, err := io.ReadAll(w.Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(body), "not authorized to delete this user")
	})
	t.Run("admin delete", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/users/tester", nil)
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusOK)
	})
	t.Run("delete non-existent user", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/users/tester3", nil)
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		response, err := io.ReadAll(w.Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(response), "user does not exist")
	})
}

func TestEditUser(t *testing.T) {
	deleteAllUsers()
	err := createTestUser(models.User{Username: "tester", Password: "testing", IsAdmin: false})
	should.BeNil(t, err)
	err = createTestUser(models.User{Username: "tester2", Password: "testing", IsAdmin: false})
	should.BeNil(t, err)
	createAdmin()

	t.Run("adminGetUser", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/tester", nil)
		should.BeNil(t, err)
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusOK)
	})

	t.Run("GetSelf", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "tester", Password: "testing"})
		should.NotBeNil(t, cookie)
		req := httptest.NewRequest(http.MethodGet, "/users/tester", nil)
		should.BeNil(t, err)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusOK)
	})

	t.Run("GetOther", func(t *testing.T) {
		w := httptest.NewRecorder()
		cookie := testLogin(models.User{Username: "tester", Password: "testing"})
		should.NotBeNil(t, cookie)
		req := httptest.NewRequest(http.MethodGet, "/users/tester2", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		response, err := io.ReadAll(w.Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(response), "non-admin cannot edit other users")
	})

	t.Run("edit other user by non-admin", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "tester", Password: "testing"})
		should.NotBeNil(t, cookie)
		w := httptest.NewRecorder()
		body := bodyParams("password", "newPassword")
		req := httptest.NewRequest(http.MethodPost, "/users/tester2", body)
		req.AddCookie(cookie)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusUnauthorized)
		response, err := io.ReadAll(w.Body)
		should.BeNil(t, err)
		should.ContainSubstring(t, string(response), "not authorized to edit this user")
	})
	t.Run("edit user by admin", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bodyParams("username", "tester2", "password", "newPassword", "admin", "on")
		req := httptest.NewRequest(http.MethodPost, "/users/tester2", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusFound)
		cookie := testLogin(models.User{Username: "tester2", Password: "newPassword"})
		should.NotBeNil(t, cookie)
	})
	t.Run("edit user by self", func(t *testing.T) {
		cookie := testLogin(models.User{Username: "tester", Password: "testing"})
		should.NotBeNil(t, cookie)

		w := httptest.NewRecorder()
		body := bodyParams("password", "newPassword")
		req := httptest.NewRequest(http.MethodPost, "/users/tester", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusFound)
		cookie = testLogin(models.User{Username: "tester", Password: "newPassword"})
		should.NotBeNil(t, cookie)
	})
	t.Run("incomplete data", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/users/tester", nil)
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		body, _ := io.ReadAll(w.Result().Body)
		should.ContainSubstring(t, string(body), "password cannot be blank")
	})
	t.Run("user does not exist", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := bodyParams("password", "newPassword")
		req := httptest.NewRequest(http.MethodPost, "/users/nosuchuser", payload)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(adminLogin())
		router.ServeHTTP(w, req)
		should.BeEqual(t, w.Code, http.StatusBadRequest)
		body, _ := io.ReadAll(w.Result().Body)
		should.ContainSubstring(t, string(body), "user does not exist")
	})
}

func testLogin(data models.User) *http.Cookie {
	w := httptest.NewRecorder()
	body := bodyParams("username", data.Username, "password", data.Password)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "devilcove-time" {
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
		_ = database.DeleteUser(user.Username)
	}
}

func createAdmin() {
	createTestUser(models.User{Username: "admin", Password: "password", IsAdmin: true})
}

func bodyParams(params ...string) io.Reader {
	body := url.Values{}
	for i := 0; i < len(params)-1; i = i + 2 {
		body.Set(params[i], params[i+1])
	}
	return strings.NewReader(body.Encode())
}

func adminLogin() *http.Cookie {
	w := httptest.NewRecorder()
	body := bodyParams("username", "admin", "password", "password")
	r := httptest.NewRequest(http.MethodPost, "/login", body)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, r)
	cookies := w.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "devilcove-time" {
			return cookie
		}
	}
	return nil
}
