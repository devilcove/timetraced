package database

import (
	"testing"

	"github.com/devilcove/timetraced/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSaveProject(t *testing.T) {
	p := models.Project{
		ID:     uuid.New(),
		Name:   "testProject",
		Active: true,
	}
	err := SaveProject(&p)
	assert.Nil(t, err)
}
