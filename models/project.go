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
	Name    string
	Active  bool
	Updated time.Time
}

type StartRequest struct {
	Project string
}

func IsTrackingActive(u string) bool {
	return trackingActive[u]
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
	return trackedProject[u]
}
