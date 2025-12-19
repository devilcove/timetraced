package models

import (
	"log/slog"
	"runtime/debug"
	"time"
)

// Page represents the data to for display to user.
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

// GetPage returns default page data.
func GetPage() Page {
	return initialize()
}

// GetUserPage returns current page for a user.
func GetUserPage(u string) Page {
	if u == "" {
		return initialize()
	}
	if page, ok := pages[u]; ok {
		return page
	}
	slog.Info("user page not set, using default")
	pages[u] = initialize()
	return pages[u]
}

// SetTheme sets the page theme for a user.
func SetTheme(user, theme string) {
	page, ok := pages[user]
	if !ok {
		page = initialize()
	}
	page.Theme = theme
	pages[user] = page
}

// SetFont sets the page font for a useruser page.
func SetFont(user, font string) {
	page, ok := pages[user]
	if !ok {
		page = initialize()
	}
	page.Font = font
	pages[user] = page
}

// SetRefresh sets the refresh rate for a user.
func SetRefresh(user string, refresh int) {
	page, ok := pages[user]
	if !ok {
		page = initialize()
	}
	page.Refresh = refresh
	pages[user] = page
}
