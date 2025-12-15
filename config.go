package main

import (
	"net/http"
	"strconv"

	"github.com/devilcove/timetraced/models"
)

func configOld(w http.ResponseWriter, r *http.Request) {
	page := models.GetPage()
	templates.ExecuteTemplate(w, "config", page)
}

func setConfig(w http.ResponseWriter, r *http.Request) {
	session := sessionData(r)
	user := session.User
	if err := r.ParseForm(); err != nil {
		processError(w, http.StatusBadRequest, "invalid data")
		return
	}
	refresh, err := strconv.Atoi(r.FormValue("refresh"))
	if err != nil {
		refresh = 5
	}
	config := models.Config{
		Theme:   r.FormValue("theme"),
		Font:    r.FormValue("font"),
		Refresh: refresh,
	}
	models.SetTheme(user, config.Theme)
	models.SetFont(user, config.Font)
	models.SetRefresh(user, config.Refresh)
	page := models.GetUserPage(user)
	w.Header().Set("HX-Refresh", "true")
	templates.ExecuteTemplate(w, "layout", page)
}
