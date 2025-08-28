package models

import (
	"time"

	"github.com/google/uuid"
)

// Report represents the time spent on a project.  It is a response to a ReportRequest.
type Report struct {
	Project string
	Total   string
	Items   []ReportRecord
}

// ReportRecord represents and individual report record.
type ReportRecord struct {
	ID    uuid.UUID
	Start time.Time
	End   time.Time
}

// ReportRequest contains data to initiate a report.
type ReportRequest struct {
	Start   string `form:"start"   json:"start"`
	End     string `form:"end"     json:"end"`
	Project string `form:"project" json:"project"`
}

// DatabaseReportRequest represents a ReportRequest formatted for db queries.
type DatabaseReportRequest struct {
	Start   time.Time
	End     time.Time
	Project string
	User    string
}
