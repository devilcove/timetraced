package models

import (
	"testing"
	"time"

	"github.com/Kairum-Labs/should"
)

func TestGetPage(t *testing.T) {
	page := GetPage()
	should.BeEqual(t, page.Theme, "indigo")
	should.BeEqual(t, page.Font, "Roboto")
	should.BeEqual(t, page.Refresh, 5)
}

func TestGetUserPage(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		page := GetUserPage("tester")
		page2 := GetUserPage("tester")
		should.BeEqual(t, page, page2)
	})
	t.Run("empty", func(t *testing.T) {
		page := GetUserPage("")
		should.BeEqual(t, page.Theme, "indigo")
	})
}

func TestSetTheme(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		page := GetUserPage("tester")
		SetTheme("tester", "red")
		page2 := GetUserPage("tester")
		should.NotBeEqual(t, page, page2)
		should.BeEqual(t, page2.Theme, "red")
	})
	t.Run("new", func(t *testing.T) {
		SetTheme("tester2", "pink")
		page := GetUserPage("tester2")
		should.BeEqual(t, page.Theme, "pink")
	})
}

func TestSetFont(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		page := GetUserPage("tester")
		SetFont("tester", "tangerine")
		page2 := GetUserPage("tester")
		should.NotBeEqual(t, page, page2)
		should.BeEqual(t, page2.Font, "tangerine")
	})
	t.Run("new", func(t *testing.T) {
		SetFont("tester3", "tangerine")
		page := GetUserPage("tester3")
		should.BeEqual(t, page.Font, "tangerine")
	})
}

func TestSetRefresh(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		page := GetUserPage("tester")
		SetRefresh("tester", 10)
		page2 := GetUserPage("tester")
		should.NotBeEqual(t, page, page2)
		should.BeEqual(t, page2.Refresh, 10)
	})
	t.Run("new", func(t *testing.T) {
		SetRefresh("tester4", 12)
		page := GetUserPage("tester4")
		should.BeEqual(t, page.Refresh, 12)
	})
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
