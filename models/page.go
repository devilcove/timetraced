package models

import (
	"runtime/debug"
	"time"
)

var version string

// Page represents the data to for display to user.
type Page struct {
	Version     string
	Tracking    bool
	Projects    []string
	Status      StatusResponse
	DefaultDate string
}

// Version reads version info from executable.
func Version() {
	version = "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		version = info.Main.Version
	}
}

// GetPage returns default page data.
func GetPage() Page {
	return Page{
		Version:     version,
		DefaultDate: time.Now().Local().Format("2006-01-02"),
	}
}
