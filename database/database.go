package database

import (
	"errors"
	"time"

	"go.etcd.io/bbolt"
)

const (
	// table names
	USERS_TABLE_NAME   = "users"
	PROJECT_TABLE_NAME = "projects"
	RECORDS_TABLE_NAME = "records"
	// sql verbs
	INIT_DB      = "init"
	CREATE_TABLE = "createtable"
	INSERT       = "insert"
	DELETE       = "delete"
	DELETE_ALL   = "deleteall"
	FETCH        = "fetch"
	CLOSE_DB     = "close"
	// errors
	NO_RECORDS = "no results found"
)

var (
	ErrNoResults = errors.New("no results found")
	db           *bbolt.DB
)

func InitializeDatabase() error {
	var err error
	db, err = bbolt.Open("./time.db", 0666, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	return createTables()
}

func Close() {
	if err := db.Close(); err != nil {
		panic(err)
	}
}

func createTables() error {
	if err := createTable(USERS_TABLE_NAME); err != nil {
		return err
	}
	if err := createTable(PROJECT_TABLE_NAME); err != nil {
		return err
	}
	if err := createTable(RECORDS_TABLE_NAME); err != nil {
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
