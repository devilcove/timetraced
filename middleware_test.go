package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kairum-Labs/should"
)

func TestAuth(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/users/", nil)
		router.ServeHTTP(w, r)
		should.BeEqual(t, w.Result().StatusCode, http.StatusSeeOther)
		should.BeEqual(t, w.Result().Header.Get("Location"), "/login/")
	})
	t.Run("authorized", func(t *testing.T) {
		createAdmin()
		r := httptest.NewRequest(http.MethodGet, "/users/", nil)
		r.AddCookie(adminLogin())
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		should.BeEqual(t, w.Result().StatusCode, http.StatusOK)
	})
}
