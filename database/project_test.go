package database

import (
	"testing"
	"time"

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

func TestGetProject(t *testing.T) {
	SaveProject(&models.Project{
		ID:      uuid.New(),
		Name:    "test",
		Active:  true,
		Updated: time.Now(),
	})
	t.Run("exists", func(t *testing.T) {
		project, err := GetProject("test")
		assert.Nil(t, err)
		assert.Equal(t, "test", project.Name)
	})
	t.Run("missing", func(t *testing.T) {
		project, err := GetProject("test2")
		assert.NotNil(t, err)
		assert.Equal(t, models.Project{}, project)
	})

}
