package database

import (
	"encoding/json"
	"time"

	"github.com/devilcove/timetraced/models"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

func SaveRecord(r *models.Record) error {
	value, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(RECORDS_TABLE_NAME))
		return b.Put([]byte(r.ID.String()), value)
	})
}

func GetRecord(id uuid.UUID) (models.Record, error) {
	record := models.Record{}
	if err := db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte(RECORDS_TABLE_NAME)).Get([]byte(id.String()))
		if err := json.Unmarshal(v, &record); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return record, err
	}
	return record, nil
}

func GetAllRecords() ([]models.Record, error) {
	var records []models.Record
	var record models.Record
	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(RECORDS_TABLE_NAME))
		_ = b.ForEach(func(k, v []byte) error {
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

func GetAllRecordsForUser(u string) ([]models.Record, error) {
	var records []models.Record
	var record models.Record
	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(RECORDS_TABLE_NAME))
		_ = b.ForEach(func(k, v []byte) error {
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

func DeleteRecord(id uuid.UUID) error {
	if err := db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(RECORDS_TABLE_NAME)).Delete([]byte(id.String()))
	}); err != nil {
		return err
	}
	return nil
}

func GetTodaysRecords() ([]models.Record, error) {
	records := []models.Record{}
	record := models.Record{}
	today := truncateToStart(time.Now())
	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(RECORDS_TABLE_NAME))
		_ = b.ForEach(func(k, v []byte) error {
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

func GetTodaysRecordsForUser(user string) ([]models.Record, error) {
	if user == "" {
		return []models.Record{}, nil
	}
	records := []models.Record{}
	record := models.Record{}
	today := truncateToStart(time.Now())
	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(RECORDS_TABLE_NAME))
		_ = b.ForEach(func(k, v []byte) error {
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

func GetReportRecords(req models.DatabaseReportRequest) ([]models.Record, error) {
	records := []models.Record{}
	record := models.Record{}
	start := truncateToStart(req.Start)
	end := truncateToEnd(req.End)
	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(RECORDS_TABLE_NAME))
		_ = b.ForEach(func(k, v []byte) error {
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
