package database

import (
	"os"
	"testing"

	"github.com/Kairum-Labs/should"
)

func TestMain(m *testing.M) {
	// main.setLogging()
	os.Setenv("DB_FILE", "test.db") //nolint:errcheck,gosec
	_ = InitializeDatabase()
	defer Close()
	// main.checkDefaultUser()
	os.Exit(m.Run())
}

func TestCloseDB(t *testing.T) {
	t.Run("open", func(t *testing.T) {
		Close()
	})
	t.Run("closed", func(t *testing.T) {
		Close()
		err := InitializeDatabase()
		should.BeNil(t, err)
	})
}
