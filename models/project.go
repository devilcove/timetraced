package models

import (
	"time"

	"github.com/google/uuid"
)

var (
	trackingActive bool
	trackedProject *string
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

func IsTrackingActive() bool {
	return trackingActive
}

func TrackingActive(p Project) {
	trackingActive = true
	trackedProject = &p.Name
}

func TrackingInactive() {
	trackingActive = false
	trackedProject = nil
}
