package database

import (
	"testing"

	"github.com/Kairum-Labs/should"
	"github.com/devilcove/timetraced/models"
	"golang.org/x/crypto/bcrypt"
)

func TestSaveUser(t *testing.T) {
	err := SaveUser(&models.User{
		Username: "testUser",
		Password: "don't care",
	})
	should.BeNil(t, err)
	should.BeNil(t, deleteAllUsers())
}

func TestDeleteUser(t *testing.T) {
	err := createTestUser(models.User{
		Username: "testUser",
		Password: "don't care",
	})
	should.BeNil(t, err)
	err = DeleteUser("testUser")
	should.BeNil(t, err)
	err = DeleteUser("testUser") // deleting a non-exitent entry does not return error
	should.BeNil(t, err)
}

func TestGetUsers(t *testing.T) {
	should.BeNil(t, deleteAllUsers())
	t.Run("no users", func(t *testing.T) {
		user, err := GetUser("testUser")
		should.NotBeNil(t, err)
		should.BeEqual(t, user, models.User{})
	})
	t.Run("one user", func(t *testing.T) {
		should.BeNil(t, createTestUser(models.User{
			Username: "testUser",
		}))
		user, err := GetUser("testUser")
		should.BeNil(t, err)
		should.BeEqual(t, user.Username, "testUser")
	})
	t.Run("multiple users", func(t *testing.T) {
		should.BeNil(t, createTestUser(models.User{
			Username: "user2",
		}))
		user, err := GetUser("testUser")
		should.BeNil(t, err)
		should.BeEqual(t, user.Username, "testUser")
	})
}

func TestGetAllUsers(t *testing.T) {
	should.BeNil(t, deleteAllUsers())
	t.Run("no users", func(t *testing.T) {
		users, err := GetAllUsers()
		should.BeNil(t, err)
		should.BeEmpty(t, users)
	})
	t.Run("one user", func(t *testing.T) {
		should.BeNil(t, createTestUser(models.User{
			Username: "testUser",
		}))
		users, err := GetAllUsers()
		should.BeNil(t, err)
		should.BeEqual(t, len(users), 1)
	})
	t.Run("multiple users", func(t *testing.T) {
		should.BeNil(t, createTestUser(models.User{
			Username: "user2",
		}))
		users, err := GetAllUsers()
		should.BeNil(t, err)
		should.BeGreaterThan(t, len(users), 1)
	})
}

func createTestUser(user models.User) error {
	user.Password, _ = hashPassword(user.Password)
	if err := SaveUser(&user); err != nil {
		return err
	}
	return nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	return string(bytes), err
}

func deleteAllUsers() error {
	users, err := GetAllUsers()
	if err != nil {
		return nil
	}
	for _, user := range users {
		if err := DeleteUser(user.Username); err != nil {
			return err
		}
	}
	return nil
}
