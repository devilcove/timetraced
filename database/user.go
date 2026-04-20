package database

import (
	"encoding/json"
	"errors"

	"github.com/devilcove/timetraced/models"
	"go.etcd.io/bbolt"
)

// SaveUser saves/updates user in db.
func SaveUser(u *models.User) error {
	value, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(userTableName))
		return b.Put([]byte(u.Username), value)
	})
}

// GetUser retrieves the named user from db.
func GetUser(name string) (models.User, error) {
	user := models.User{}
	if err := db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte(userTableName)).Get([]byte(name))
		if v == nil {
			return errors.New("no such user")
		}
		if err := json.Unmarshal(v, &user); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return user, err
	}
	return user, nil
}

// GetAllUsers retrieves all users from db.
func GetAllUsers() ([]models.User, error) {
	var users []models.User
	var user models.User
	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(userTableName))
		if b == nil {
			return errors.New("no users")
		}
		_ = b.ForEach(func(_, v []byte) error {
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

// DeleteUser deletes a user from db.
func DeleteUser(name string) error {
	if err := db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(userTableName)).Delete([]byte(name))
	}); err != nil {
		return err
	}
	return nil
}
