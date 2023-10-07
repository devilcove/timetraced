package database

import (
	"encoding/json"

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
		b.ForEach(func(k, v []byte) error {
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

func DeleteRecord(id uuid.UUID) error {
	if err := db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(RECORDS_TABLE_NAME)).Delete([]byte(id.String()))
	}); err != nil {
		return err
	}
	return nil
}
