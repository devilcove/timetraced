package database

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	//main.setLogging()
	os.Setenv("DB_FILE", "test.db") //nolint:errcheck
	_ = InitializeDatabase()
	defer Close()
	//main.checkDefaultUser()
	os.Exit(m.Run())
}

func TestCloseDB(t *testing.T) {
	t.Run("open", func(t *testing.T) {
		Close()
	})
	t.Run("closed", func(t *testing.T) {
		Close()
		err := InitializeDatabase()
		assert.Nil(t, err)
	})
}
