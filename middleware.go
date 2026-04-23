package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/devilcove/cookie"
	"github.com/devilcove/timetraced/models"
)

type contextKey string

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getCookie(r)
		if user.Username == "" {
			slog.Error("unauthorized", "user", user)
			w.WriteHeader(http.StatusUnauthorized)
			render(w, "loginForm", nil)
			return
		}
		ctx := context.WithValue(r.Context(), contextKey("user"), user)
		saveCookie(user, w) // refresh cookie
		slog.Info("auth: set user context", "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getRequestUser(r *http.Request) models.User {
	user, ok := r.Context().Value(contextKey("user")).(models.User)
	if !ok {
		slog.Debug("get user from context", "value", r.Context().Value(contextKey("user")))
	}
	return user
}

func saveCookie(user models.User, w http.ResponseWriter) {
	user.Password = ""
	bytes, err := json.Marshal(user)
	if err != nil {
		slog.Error("marshal cookie", "error", err)
		return
	}
	if err := cookie.Save(w, cookieName, bytes); err != nil {
		slog.Error("save cookie", "error", err)
	}
}

func getCookie(r *http.Request) models.User {
	var user models.User
	data, err := cookie.Get(r, cookieName)
	if err != nil {
		slog.Error("get cookie", "error", err)
		return user
	}
	if err := json.Unmarshal(data, &user); err != nil {
		slog.Error("unmarshal cookie", "error", err)
	}
	return user
}
