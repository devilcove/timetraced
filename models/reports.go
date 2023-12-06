package models

import (
	"time"

	"github.com/google/uuid"
)

type Report struct {
	Project string
	Total   string
	Items   []ReportRecord
}

type ReportRecord struct {
	ID    uuid.UUID
	Start time.Time
	End   time.Time
}

type ReportRequest struct {
	Start   string `form:"start"`
	End     string `form:"end"`
	Project string `form:"project"`
}

type DatabaseReportRequest struct {
	Start   time.Time
	End     time.Time
	Project string
	User    string
}
