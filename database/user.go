package database

import (
	"encoding/json"

	"github.com/devilcove/timetraced/models"
	"go.etcd.io/bbolt"
)

func SaveUser(u *models.User) error {
	value, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(USERS_TABLE_NAME))
		return b.Put([]byte(u.Username), value)
	})
}

func GetUser(name string) (models.User, error) {
	user := models.User{}
	if err := db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte(USERS_TABLE_NAME)).Get([]byte(name))
		if err := json.Unmarshal(v, &user); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return user, err
	}
	return user, nil
}

func GetAllUsers() ([]models.User, error) {
	var users []models.User
	var user models.User
	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(USERS_TABLE_NAME))
		b.ForEach(func(k, v []byte) error {
			if err := json.Unmarshal(v, &user); err != nil {
				return err
			}
			users = append(users, user)
			return nil
		})
		return nil
	}); err != nil {
		return users, err
	}
	return users, nil
}

func DeleteUser(name string) error {
	if err := db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(USERS_TABLE_NAME)).Delete([]byte(name))
	}); err != nil {
		return err
	}
	return nil
}
