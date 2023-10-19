package database

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	//main.setLogging()
	os.Setenv("DB_FILE", "test.db")
	InitializeDatabase()
	defer Close()
	//main.checkDefaultUser()
	os.Exit(m.Run())
}
