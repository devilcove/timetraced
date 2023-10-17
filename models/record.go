package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Record struct {
	ID      uuid.UUID
	Project string
	User    string
	Start   time.Time
	End     time.Time
}

type Durations map[string]string

type Duration struct {
	Project string
	Elapsed string
}

func (r *Record) Duration() time.Duration {
	return r.End.Sub(r.Start)
}

func FmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	dm := int(m) / 6
	if m%6 != 0 {
		dm += 1
	}
	dh := h
	if dm == 10 {
		dh += 1
		dm = 0
	}
	return fmt.Sprintf("%02d:%02d (%2d.%d Hours)", h, m, dh, dm)
}

type StatusResponse struct {
	Current      string
	Elapsed      string
	CurrentTotal string
	DailyTotal   string
	Durations    []Duration
}

type Status struct {
	Current    string
	Elapsed    time.Duration
	Total      time.Duration
	DailyTotal time.Duration
}
