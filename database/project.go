package database

import (
	"encoding/json"
	"errors"

	"github.com/devilcove/timetraced/models"
	"go.etcd.io/bbolt"
)

// SaveProject saves a project to db.
func SaveProject(p *models.Project) error {
	value, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(projectTableName))
		return b.Put([]byte(p.Name), value)
	})
}

// GetProject retrives a project from db.
func GetProject(name string) (models.Project, error) {
	project := models.Project{}
	if err := db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte(projectTableName)).Get([]byte(name))
		if v == nil {
			return errors.New("no such project")
		}
		if err := json.Unmarshal(v, &project); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return project, err
	}
	return project, nil
}

// GetAllProjects retrieves all projects from db.
func GetAllProjects() ([]models.Project, error) {
	var projects []models.Project
	var project models.Project
	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(projectTableName))
		_ = b.ForEach(func(_, v []byte) error {
			if err := json.Unmarshal(v, &project); err != nil {
				return err
			}
			projects = append(projects, project)
			return nil
		})
		return nil
	}); err != nil {
		return projects, err
	}
	return projects, nil
}

// DeleteProject deletes a project from the db.
func DeleteProject(name string) error {
	if err := db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(projectTableName)).Delete([]byte(name))
	}); err != nil {
		return err
	}
	return nil
}

// GetActiveProject retrieves the project for which time is aatively being recorded.
func GetActiveProject(u string) *models.Project {
	records, err := GetTodaysRecords()
	if err != nil {
		return nil
	}
	for _, record := range records {
		if record.User != u {
			continue
		}
		if record.End.IsZero() {
			project, err := GetProject(record.Project)
			if err != nil {
				return nil
			}
			return &project
		}
	}
	return nil
}
