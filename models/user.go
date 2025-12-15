package models

import (
	"time"
)

// User represents a user.
type User struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
	IsAdmin  bool
	Updated  time.Time
}

type Editor struct {
	User
	AsAdmin bool
}
