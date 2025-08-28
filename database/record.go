package database

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/devilcove/timetraced/models"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

// SaveRecord saves a record to the db.
func SaveRecord(r *models.Record) error {
	value, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(recordsTableName))
		return b.Put([]byte(r.ID.String()), value)
	})
}

// GetRecord retrives a record form db.
func GetRecord(id uuid.UUID) (models.Record, error) {
	record := models.Record{}
	if err := db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte(recordsTableName)).Get([]byte(id.String()))
		if v == nil {
			return errors.New("no such record")
		}
		if err := json.Unmarshal(v, &record); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return record, err
	}
	return record, nil
}

// GetAllRecords returns all records from db.
func GetAllRecords() ([]models.Record, error) {
	var records []models.Record
	var record models.Record
	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(recordsTableName))
		_ = b.ForEach(func(_, v []byte) error {
			if err := json.Unmarshal(v, &record); err != nil {
				return err
			}
			records = append(records, record)
			return nil
		})
		return nil
	}); err != nil {
		return records, err
	}
	return records, nil
}

// GetAllRecordsForUser returns all records created by user from db.
func GetAllRecordsForUser(u string) ([]models.Record, error) {
	var records []models.Record
	var record models.Record
	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(recordsTableName))
		_ = b.ForEach(func(_, v []byte) error {
			if err := json.Unmarshal(v, &record); err != nil {
				return err
			}
			if record.User == u {
				records = append(records, record)
			}
			return nil
		})
		return nil
	}); err != nil {
		return records, err
	}
	return records, nil
}

// DeleteRecord deletes a record from db.
func DeleteRecord(id uuid.UUID) error {
	if err := db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(recordsTableName)).Delete([]byte(id.String()))
	}); err != nil {
		return err
	}
	return nil
}

// GetTodaysRecords returns records created on this day.
func GetTodaysRecords() ([]models.Record, error) {
	records := []models.Record{}
	record := models.Record{}
	today := truncateToStart(time.Now())
	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(recordsTableName))
		_ = b.ForEach(func(_, v []byte) error {
			if err := json.Unmarshal(v, &record); err != nil {
				return err
			}
			if record.Start.After(today) {
				records = append(records, record)
			}
			return nil
		})
		return nil
	}); err != nil {
		return records, err
	}
	return records, nil
}

// GetTodaysRecordsForUser return records created on this day by specified user.
func GetTodaysRecordsForUser(user string) ([]models.Record, error) {
	if user == "" {
		return []models.Record{}, nil
	}
	records := []models.Record{}
	record := models.Record{}
	today := truncateToStart(time.Now())
	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(recordsTableName))
		_ = b.ForEach(func(_, v []byte) error {
			if err := json.Unmarshal(v, &record); err != nil {
				return err
			}
			if record.User == user && record.Start.After(today) {
				records = append(records, record)
			}
			return nil
		})
		return nil
	}); err != nil {
		return records, err
	}
	return records, nil
}

// GetReportRecords returns record matching the request.
func GetReportRecords(req models.DatabaseReportRequest) ([]models.Record, error) {
	records := []models.Record{}
	record := models.Record{}
	start := truncateToStart(req.Start)
	end := truncateToEnd(req.End)
	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(recordsTableName))
		_ = b.ForEach(func(_, v []byte) error {
			if err := json.Unmarshal(v, &record); err != nil {
				return err
			}
			if (req.User == record.User) &&
				(req.Project == record.Project) &&
				record.Start.After(start) &&
				record.Start.Before(end) {
				if record.End.IsZero() {
					record.End = time.Now()
				}
				records = append(records, record)
			}
			return nil
		})
		return nil
	}); err != nil {
		return records, err
	}
	return records, nil
}

func truncateToStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func truncateToEnd(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}
