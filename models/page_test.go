package models

import (
	"testing"
	"time"

	"github.com/Kairum-Labs/should"
)

func TestGetPage(t *testing.T) {
	Version()
	page := GetPage()
	should.BeEqual(t, page.Version, "(devel)")
	t.Log(page)
}

func TestTracking(t *testing.T) {
	tracking := IsTrackingActive("tester")
	should.BeFalse(t, tracking)
	project := Tracked("tester")
	should.BeEqual(t, project, "")
	TrackingActive("tester", Project{Name: "project"})
	tracking = IsTrackingActive("tester")
	should.BeTrue(t, tracking)
	project = Tracked("tester")
	should.BeEqual(t, project, "project")
	TrackingInactive("tester")
	tracking = IsTrackingActive("tester")
	should.BeFalse(t, tracking)
}

func TestRecords(t *testing.T) {
	s := FmtDuration(time.Minute * 57)
	should.BeEqual(t, s, "00:57 ( 1.0 Hours)")
	record := Record{
		End:   time.Now(),
		Start: time.Now().Add(time.Hour * -1),
	}
	expected := record.End.Sub(record.Start)
	dur := record.Duration()
	should.BeEqual(t, dur, expected)
}
