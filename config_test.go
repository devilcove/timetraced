package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kairum-Labs/should"
)

func TestConfig(t *testing.T) {
	createAdmin()
	c := adminLogin()
	if c == nil {
		t.FailNow()
	}
	t.Run("get", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/config/", nil)
		r.AddCookie(adminLogin())
		router.ServeHTTP(w, r)
		should.BeEqual(t, w.Result().StatusCode, http.StatusOK)
	})
	t.Run("update", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bodyParams("theme", "red", "font", "tangerine", "refesh", "10")
		r := httptest.NewRequest(http.MethodPost, "/config/", body)
		r.AddCookie(adminLogin())
		router.ServeHTTP(w, r)
		should.BeEqual(t, w.Result().StatusCode, http.StatusOK)
	})
	t.Run("invalid refresh", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bodyParams("theme", "red", "font", "tangerine", "refesh", "junk")
		r := httptest.NewRequest(http.MethodPost, "/config/", body)
		r.AddCookie(adminLogin())
		router.ServeHTTP(w, r)
		should.BeEqual(t, w.Result().StatusCode, http.StatusOK)

	})
}
