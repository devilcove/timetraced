package main

import (
	"testing"

	"github.com/Kairum-Labs/should"
	"github.com/devilcove/timetraced/database"
)

func TestDefaultUser(t *testing.T) {
	deleteAllUsers()
	users, err := database.GetAllUsers()
	should.BeNil(t, err)
	checkDefaultUser()
	users, err = database.GetAllUsers()
	should.BeNil(t, err)
	should.BeEqual(t, len(users), 1)
	should.BeEqual(t, users[0].Username, "admin")
	checkDefaultUser() // run second time
	users, err = database.GetAllUsers()
	should.BeNil(t, err)
	should.BeEqual(t, len(users), 1)
	should.BeEqual(t, users[0].Username, "admin")
}
