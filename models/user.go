package models

import (
	"time"
)

type User struct {
	Username string
	Password string
	IsAdmin  bool
	Updated  time.Time
}
