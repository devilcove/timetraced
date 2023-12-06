package models

import (
	"log"
	"time"
)

type Page struct {
	Page        string
	Version     string
	Theme       string
	Font        string
	Tracking    bool
	Projects    []string
	Status      StatusResponse
	DefaultDate string
}

// var page Page
var pages map[string]Page

func initialize() Page {
	return Page{
		Page:        "login",
		Version:     "v0.1.0",
		Theme:       "indigo",
		Font:        "Roboto",
		DefaultDate: time.Now().Local().Format("2006-01-02"),
	}
}
func init() {
	pages = make(map[string]Page)
}

func GetPage() Page {
	return initialize()
}

func GetUserPage(u string) Page {
	if page, ok := pages[u]; ok {
		return page
	}
	log.Println("user page not set, using default")
	pages[u] = initialize()
	return pages[u]
}

func SetTheme(user, theme string) {
	page, ok := pages[user]
	if !ok {
		page = initialize()
	}
	page.Theme = theme
	pages[user] = page
}

func SetFont(user, font string) {
	page, ok := pages[user]
	if !ok {
		page = initialize()
	}
	page.Font = font
	pages[user] = page
}

func SetPage(user, p string) {
	page, ok := pages[user]
	if !ok {
		page = initialize()
	}
	page.Page = p
	pages[user] = page
}
