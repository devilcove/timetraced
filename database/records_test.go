package database

import (
	"testing"
	"time"

	"github.com/Kairum-Labs/should"
	"github.com/devilcove/timetraced/models"
	"github.com/google/uuid"
)

func Test_truncateToStart(t *testing.T) {
	tests := []struct {
		name string
		args time.Time
		want time.Time
	}{
		{
			name: "UTC",
			args: time.Date(2023, 1, 1, 17, 34, 59, 0, time.UTC),
			want: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "local",
			args: time.Date(1918, 10, 30, 12, 1, 9, 0, time.Local),
			want: time.Date(1918, 10, 30, 0, 0, 0, 0, time.Local),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateToStart(tt.args); got.Compare(tt.want) != 0 {
				t.Errorf("truncateToStart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_truncateToEnd(t *testing.T) {
	tests := []struct {
		name string
		args time.Time
		want time.Time
	}{
		{
			name: "UTC",
			args: time.Date(2023, 1, 1, 17, 34, 59, 0, time.UTC),
			want: time.Date(2023, 1, 1, 23, 59, 59, 0, time.UTC),
		},
		{
			name: "local",
			args: time.Date(1918, 10, 30, 12, 1, 9, 0, time.Local),
			want: time.Date(1918, 10, 30, 23, 59, 59, 0, time.Local),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateToEnd(tt.args); got.Compare(tt.want) != 0 {
				t.Errorf("truncateToStart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSaveRecord(t *testing.T) {
	should.BeNil(t, deleteAllRecords())
	err := SaveRecord(&models.Record{
		ID:      uuid.New(),
		Project: "one",
		User:    "testUser",
		Start:   time.Now().Add(time.Hour * -1),
		End:     time.Now(),
	})
	should.BeNil(t, err)
}

func TestGetRecord(t *testing.T) {
	should.BeNil(t, deleteAllRecords())
	should.BeNil(t, createTestRecords())
	records, err := GetAllRecords()
	should.BeNil(t, err)
	should.BeEqual(t, len(records), 3)
	record, err := GetRecord(records[0].ID)
	should.BeNil(t, err)
	should.BeEqual(t, record.User, records[0].User)
	//t.Log(record)
}

func TestGetTodaysRecords(t *testing.T) {
	should.BeNil(t, deleteAllRecords())
	should.BeNil(t, createTestRecords())
	t.Run("all", func(t *testing.T) {
		records, err := GetTodaysRecords()
		should.BeNil(t, err)
		should.BeEqual(t, len(records), 3)
	})
	t.Run("forUser", func(t *testing.T) {
		records, err := GetTodaysRecordsForUser("testUser")
		should.BeNil(t, err)
		should.BeEqual(t, len(records), 2)
	})
}

func TestGetReportRecords(t *testing.T) {
	should.BeNil(t, deleteAllRecords())
	should.BeNil(t, createTestRecords())
	t.Run("today", func(t *testing.T) {
		records, err := GetReportRecords(models.DatabaseReportRequest{
			Start:   time.Now(),
			End:     time.Now(),
			Project: "one",
			User:    "testUser",
		})
		should.BeNil(t, err)
		should.BeEqual(t, len(records), 2)
	})
	t.Run("yesterday", func(t *testing.T) {
		records, err := GetReportRecords(models.DatabaseReportRequest{
			Start:   time.Now().Add(time.Hour * -24 * 7),
			End:     time.Now().Add(time.Hour * -24),
			Project: "one",
			User:    "testUser",
		})
		should.BeNil(t, err)
		should.BeEqual(t, len(records), 0)
	})
}

func TestDeleteRecords(t *testing.T) {
	should.BeNil(t, deleteAllRecords())
	should.BeNil(t, createTestRecords())
	records, err := GetAllRecords()
	should.BeNil(t, err)
	err = DeleteRecord(records[0].ID)
	should.BeNil(t, err)
	remainder, err := GetAllRecords()
	should.BeNil(t, err)
	should.BeLessThan(t, len(remainder), len(records))
}

func TestGetAllRecordsForUser(t *testing.T) {
	should.BeNil(t, deleteAllRecords())
	should.BeNil(t, createTestRecords())
	records, err := GetAllRecordsForUser("testUser")
	should.BeNil(t, err)
	should.BeEqual(t, len(records), 2)
}

func createTestRecords() error {
	records := []models.Record{
		{
			ID:      uuid.New(),
			Project: "one",
			User:    "testUser",
			Start:   time.Now().Add(time.Hour * -1),
			//End:     time.Now(),
		},
		{
			ID:      uuid.New(),
			Project: "one",
			User:    "testUser",
			Start:   time.Now().Add(time.Hour * -2),
			End:     time.Now().Add(time.Hour * -1),
		},
		{
			ID:      uuid.New(),
			Project: "two",
			User:    "user1",
			Start:   time.Now().Add(time.Hour * -2),
			End:     time.Now().Add(time.Hour * -1),
		},
	}
	for _, record := range records {
		if err := SaveRecord(&record); err != nil {
			return err
		}
	}
	return nil
}

func deleteAllRecords() error {
	records, err := GetAllRecords()
	if err != nil {
		return err
	}
	for _, record := range records {
		if err := DeleteRecord(record.ID); err != nil {
			return err
		}
	}
	return nil
}
