package main

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
)

// Session represents a user session.
type Session struct {
	User     string
	LoggedIn bool
	Admin    bool
	Session  *sessions.Session
}

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := sessionData(r)
		if session == nil {
			slog.Error("nil session data")
			http.Redirect(w, r, "/login/", http.StatusSeeOther)
			return
		}
		slog.Debug("auth", "session", session)
		if !session.LoggedIn {
			slog.Error("not logged in")
			http.Redirect(w, r, "/login/", http.StatusSeeOther)
			return
		}
		if err := session.Session.Save(r, w); err != nil {
			slog.Error("save session", "error", err)
		}
		next.ServeHTTP(w, r)
	})
}

func sessionData(r *http.Request) *Session {
	sess := &Session{}
	session, err := store.Get(r, "devilcove-time")
	if err != nil {
		slog.Error("session err", "error", err)
		return nil
	}
	user := session.Values["user"]
	loggedIn := session.Values["loggedIn"]
	admin := session.Values["admin"]
	if x, ok := loggedIn.(bool); ok {
		sess.LoggedIn = x
	}
	if u, ok := user.(string); ok {
		sess.User = u
	}
	if a, ok := admin.(bool); ok {
		sess.Admin = a
	}
	sess.Session = session
	return sess
}
