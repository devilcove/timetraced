package main

import (
	"net/http"
	"strconv"

	"github.com/devilcove/timetraced/models"
)

func configOld(w http.ResponseWriter, _ *http.Request) {
	page := models.GetPage()
	render(w, "config", page)
}

func setConfig(w http.ResponseWriter, r *http.Request) {
	user := getRequestUser(r)
	refresh, err := strconv.Atoi(r.FormValue("refresh"))
	if err != nil {
		refresh = 5
	}
	config := models.Config{
		Theme:   r.FormValue("theme"),
		Font:    r.FormValue("font"),
		Refresh: refresh,
	}
	models.SetTheme(user.Username, config.Theme)
	models.SetFont(user.Username, config.Font)
	models.SetRefresh(user.Username, config.Refresh)
	page := populatePage(user.Username)
	render(w, "content", page)
}
