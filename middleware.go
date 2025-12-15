package main

import (
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
	logger.Debug("auth")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := sessionData(r)
		if session == nil {
			logger.Error("nil session data")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if !session.LoggedIn {
			logger.Error("not logged in")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if err := session.Session.Save(r, w); err != nil {
			logger.Error("save session", "error", err)
		}
		next.ServeHTTP(w, r)
	})
}

func sessionData(r *http.Request) *Session {
	sess := &Session{}
	session, err := store.Get(r, "devilcove-time")
	if err != nil {
		logger.Error("session err", "error", err)
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
	logger.Debug("session", "data", sess)
	return sess
}
