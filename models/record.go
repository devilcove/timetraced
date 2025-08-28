package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Record represents a time record.
type Record struct {
	ID      uuid.UUID
	Project string
	User    string
	Start   time.Time
	End     time.Time
}

// EditRecord represents a time record for editing in UI.
type EditRecord struct {
	ID        string
	Start     string
	StartTime string
	End       string
	EndTime   string
}

// type Durations map[string]string

// Duration reprents the time spend on a project.
type Duration struct {
	Project string
	Elapsed string
}

// Duration returns the elapsed time from a time record.
func (r *Record) Duration() time.Duration {
	return r.End.Sub(r.Start)
}

// FmtDuration returns a human readable representation of a duration
// in hours:minute and decimal hours format.
func FmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour //nolint:durationcheck
	m := d / time.Minute
	dm := int(m) / 6
	if m%6 != 0 {
		dm++
	}
	dh := h
	if dm == 10 {
		dh++
		dm = 0
	}
	return fmt.Sprintf("%02d:%02d (%2d.%d Hours)", h, m, dh, dm)
}

// StatusResponse provides a status for display.
type StatusResponse struct {
	Current      string
	Elapsed      string
	CurrentTotal string
	DailyTotal   string
	Durations    []Duration
}

// Status represents the time recorded today.
type Status struct {
	Current    string
	Elapsed    time.Duration
	Total      time.Duration
	DailyTotal time.Duration
}
