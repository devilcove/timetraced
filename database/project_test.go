package database

import (
	"testing"
	"time"

	"github.com/Kairum-Labs/should"
	"github.com/devilcove/timetraced/models"
	"github.com/google/uuid"
)

func TestSaveProject(t *testing.T) {
	p := models.Project{
		ID:     uuid.New(),
		Name:   "testProject",
		Active: true,
	}
	err := SaveProject(&p)
	should.BeNil(t, err)
}

func TestGetProject(t *testing.T) {
	err := SaveProject(&models.Project{
		ID:      uuid.New(),
		Name:    "test",
		Active:  true,
		Updated: time.Now(),
	})
	should.BeNil(t, err)
	t.Run("exists", func(t *testing.T) {
		project, err := GetProject("test")
		should.BeNil(t, err)
		should.BeEqual(t, project.Name, "test")
	})
	t.Run("missing", func(t *testing.T) {
		project, err := GetProject("test2")
		should.NotBeNil(t, err)
		should.BeEqual(t, project, models.Project{})
	})
}
