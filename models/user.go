package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string
	Password string
	IsAdmin  bool
	Updated  time.Time
}

func (a *User) IsValidPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	return err == nil
}
