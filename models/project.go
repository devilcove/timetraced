package models

import (
	"time"

	"github.com/google/uuid"
)

var (
	trackingActive map[string]bool
	trackedProject map[string]string
)

type Project struct {
	ID      uuid.UUID
	Name    string `json:"name" form:"name"`
	Active  bool
	Updated time.Time
}

type StartRequest struct {
	Project string
}

func init() {
	trackingActive = make(map[string]bool)
	trackedProject = make(map[string]string)
}

func IsTrackingActive(u string) bool {
	if active, ok := trackingActive[u]; ok {
		return active
	}
	return false
}

func TrackingActive(u string, p Project) {
	trackingActive[u] = true
	trackedProject[u] = p.Name
}

func TrackingInactive(u string) {
	trackingActive[u] = false
	trackedProject[u] = ""
}

func Tracked(u string) string {
	if project, ok := trackedProject[u]; ok {
		return project
	}
	return ""
}
