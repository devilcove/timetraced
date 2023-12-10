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
	Start   string `json:"start" form:"start"`
	End     string `json:"end" form:"end"`
	Project string `json:"project" form:"project"`
}

type DatabaseReportRequest struct {
	Start   time.Time
	End     time.Time
	Project string
	User    string
}
