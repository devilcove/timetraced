package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kr/pretty"
	"golang.org/x/crypto/bcrypt"
)

const SessionAge = 60 * 60 * 8 // 8 hours in seconds

func Login(c *gin.Context) {
	var user models.User
	if err := c.BindJSON(&user); err != nil {
		ProcessError(c, http.StatusBadRequest, "invalid user")
		log.Println("bind err", err)
		return
	}
	log.Println("login by", user)
	if !ValidateUser(&user) {
		ProcessError(c, http.StatusBadRequest, "invalid user")
		log.Println("validation error")
		return
	}
	session := sessions.Default(c)
	session.Set("loggedin", true)
	session.Set("user", user.Username)
	session.Options(sessions.Options{MaxAge: SessionAge, Secure: true, SameSite: http.SameSiteLaxMode})
	session.Save()
	//location := url.URL{Path: "/"}
	//c.Redirect(http.StatusFound, location.RequestURI())
	c.Status(http.StatusNoContent)
}

func ValidateUser(visitor *models.User) bool {
	user, err := database.GetUser(visitor.Username)

	if err != nil {
		return false
	}
	if visitor.Username == user.Username && CheckPassword(visitor, &user) {
		return true
	}
	return false
}

func CheckPassword(plain, hash *models.User) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash.Password), []byte(plain.Password))
	if err != nil {
		log.Println("bcrypt", err)
	}
	return err == nil
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	//delete cookie
	session.Options(sessions.Options{MaxAge: -1})
	session.Clear()
	session.Save()
	c.Status(http.StatusNoContent)
	//location := url.URL{Path: "/"}
	//c.Redirect(http.StatusFound, location.RequestURI())
}

func New(c *gin.Context) {
	var user models.User
	var err error
	if err := c.BindJSON(&user); err != nil {
		ProcessError(c, http.StatusBadRequest, "could not decode request into json")
		return
	}
	users, err := database.GetAllUsers()
	pretty.Println(err, users)
	if err == nil {
		ProcessError(c, http.StatusBadRequest, "user exists")
		return
	}
	pretty.Println(user)
	if user.Username == "" || user.Password == "" {
		ProcessError(c, http.StatusBadRequest, "username or password cannot be blank")
		return
	}
	user.Password, err = hashPassword(user.Password)
	user.ID = uuid.New()
	if err != nil {
		ProcessError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if err := database.SaveUser(&user); err != nil {
		ProcessError(c, http.StatusInternalServerError, err.Error())
		return
	}
	location := url.URL{Path: "/"}
	c.Redirect(http.StatusFound, location.RequestURI())
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	return string(bytes), err
}
