package database

import (
	"encoding/json"

	"github.com/devilcove/timetraced/models"
	"go.etcd.io/bbolt"
)

func SaveProject(p *models.Project) error {
	value, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(PROJECT_TABLE_NAME))
		return b.Put([]byte(p.Name), value)
	})
}

func GetProject(name string) (models.Project, error) {
	project := models.Project{}
	if err := db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte(PROJECT_TABLE_NAME)).Get([]byte(name))
		if err := json.Unmarshal(v, &project); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return project, err
	}
	return project, nil
}

func GetAllProjects() ([]models.Project, error) {
	var projects []models.Project
	var project models.Project
	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(PROJECT_TABLE_NAME))
		b.ForEach(func(k, v []byte) error {
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

func DeleteProject(name string) error {
	if err := db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(PROJECT_TABLE_NAME)).Delete([]byte(name))
	}); err != nil {
		return err
	}
	return nil
}
