package database

import (
	"errors"
	"os"
	"time"

	"go.etcd.io/bbolt"
)

const (
	// table names.
	userTableName    = "users"
	projectTableName = "projects"
	recordsTableName = "records"
)

var (
	// ErrNoResults is returned when a db record does not exist in db.
	ErrNoResults = errors.New("no results found")
	db           *bbolt.DB
)

// InitializeDatabase opens (creates if it does not exist) the db and creates any non-exitent tables.
func InitializeDatabase() error {
	var err error
	file := os.Getenv("DB_FILE")
	if file == "" {
		file = "time.db"
	}
	db, err = bbolt.Open(file, 0o666, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	return createTables()
}

// Close closes the db file.
func Close() {
	if err := db.Close(); err != nil {
		panic(err)
	}
}

func createTables() error {
	if err := createTable(userTableName); err != nil {
		return err
	}
	if err := createTable(projectTableName); err != nil {
		return err
	}
	if err := createTable(recordsTableName); err != nil {
		return err
	}
	return nil
}

func createTable(name string) error {
	if err := db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(name)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
