package models

import (
	"log"
	"runtime/debug"
	"time"
)

type Page struct {
	NeedsLogin  bool
	Version     string
	Theme       string
	Font        string
	Refresh     int
	Tracking    bool
	Projects    []string
	Status      StatusResponse
	DefaultDate string
}

// var page Page
var pages map[string]Page

func initialize() Page {
	return Page{
		Version:     version(),
		Theme:       "indigo",
		Font:        "Roboto",
		Refresh:     5,
		DefaultDate: time.Now().Local().Format("2006-01-02"),
	}
}
func init() {
	pages = make(map[string]Page)
}

func version() string {
	version := "v0.1.0"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return version + " " + setting.Value
			}
		}
	}
	return version
}

func GetPage() Page {
	return initialize()
}

func GetUserPage(u string) Page {
	if u == "" {
		return initialize()
	}
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

func SetRefresh(user string, refresh int) {
	page, ok := pages[user]
	if !ok {
		page = initialize()
	}
	page.Refresh = refresh
	pages[user] = page
}
