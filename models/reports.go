package models

import "time"

type Report struct {
	Project string
	Total   time.Duration
	Items   []ReportRecord
}

type ReportRecord struct {
	Start time.Time
	End   time.Time
}

type ReportRequest struct {
	Start    time.Time
	End      time.Time
	Projects []string
	Users    []string
}
