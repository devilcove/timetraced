package models

import (
	"time"

	"github.com/google/uuid"
)

var (
	trackingActive map[string]bool
	trackedProject map[string]string
)

// Project represents a project.
type Project struct {
	ID      uuid.UUID
	Name    string
	Active  bool
	Updated time.Time
}

// StartRequest is a request to start recording time for a given project.
type StartRequest struct {
	Project string
}

func init() {
	trackingActive = make(map[string]bool)
	trackedProject = make(map[string]string)
}

// IsTrackingActive checks if tracking has been activated for given user.
func IsTrackingActive(u string) bool {
	if active, ok := trackingActive[u]; ok {
		return active
	}
	return false
}

// TrackingActive activates tracking for given user and project.
func TrackingActive(u string, p Project) {
	trackingActive[u] = true
	trackedProject[u] = p.Name
}

// TrackingInactive deactivates tracking for given user.
func TrackingInactive(u string) {
	trackingActive[u] = false
	trackedProject[u] = ""
}

// Tracked returns the poject being tracked for user.
func Tracked(u string) string {
	if project, ok := trackedProject[u]; ok {
		return project
	}
	return ""
}
