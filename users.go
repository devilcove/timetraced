package main

import (
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/kr/pretty"
	"golang.org/x/crypto/bcrypt"
)

const SessionAge = 60 * 60 * 8 // 8 hours in seconds

func login(c *gin.Context) {
	session := sessions.Default(c)
	var user models.User
	if err := c.BindJSON(&user); err != nil {
		processError(c, http.StatusBadRequest, "invalid user")
		log.Println("bind err", err)
		return
	}
	log.Println("login by", user)
	if !validateUser(&user) {
		session.Clear()
		session.Save()
		processError(c, http.StatusBadRequest, "invalid user")
		log.Println("validation error")
		return
	}
	session.Set("loggedin", true)
	session.Set("user", user.Username)
	session.Set("admin", true)
	session.Options(sessions.Options{MaxAge: SessionAge, Secure: true, SameSite: http.SameSiteLaxMode})
	session.Save()
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

func validateUser(visitor *models.User) bool {
	user, err := database.GetUser(visitor.Username)

	if err != nil {
		slog.Error("no such user", "user", visitor.Username, "error", err)
		return false
	}
	if visitor.Username == user.Username && checkPassword(visitor, &user) {
		return true
	}
	return false
}

func checkPassword(plain, hash *models.User) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash.Password), []byte(plain.Password))
	if err != nil {
		log.Println("bcrypt", err)
	}
	return err == nil
}

func logout(c *gin.Context) {
	session := sessions.Default(c)
	//delete cookie
	session.Clear()
	session.Save()
	c.Status(http.StatusNoContent)
}

func addUser(c *gin.Context) {
	var user models.User
	var err error
	session := sessions.Default(c)
	admin := session.Get("admin")
	if !admin.(bool) {
		processError(c, http.StatusUnauthorized, "only admins can create new users")
	}
	if err := c.BindJSON(&user); err != nil {
		processError(c, http.StatusBadRequest, "could not decode request into json")
		return
	}
	users, err := database.GetUser(user.Username)
	pretty.Println(err, users)
	if err == nil {
		processError(c, http.StatusBadRequest, "user exists")
		return
	}
	pretty.Println(user)
	if user.Username == "" || user.Password == "" {
		processError(c, http.StatusBadRequest, "username or password cannot be blank")
		return
	}
	user.Password, err = hashPassword(user.Password)
	if err != nil {
		processError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if err := database.SaveUser(&user); err != nil {
		processError(c, http.StatusInternalServerError, err.Error())
		return
	}
	location := url.URL{Path: "/"}
	c.Redirect(http.StatusFound, location.RequestURI())
}

func editUser(c *gin.Context) {
	var user models.User
	var err error
	session := sessions.Default(c)
	admin := session.Get("admin")
	visitor := session.Get("user")
	if err := c.BindJSON(&user); err != nil {
		processError(c, http.StatusBadRequest, "could not decode request into json")
		return
	}
	if user.Username != visitor && !admin.(bool) {
		processError(c, http.StatusUnauthorized, "you are not authorized to edit this user")
	}
	updatedUser, err := database.GetUser(user.Username)
	if err != nil {
		processError(c, http.StatusBadRequest, "user does not exists")
		return
	}
	updatedUser.Password, err = hashPassword(user.Password)
	if err != nil {
		processError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if !admin.(bool) {
		updatedUser.IsAdmin = false
	}
	updatedUser.Updated = time.Now()
	if err := database.SaveUser(&user); err != nil {
		processError(c, http.StatusInternalServerError, err.Error())
		return
	}
	updatedUser.Password = ""
	c.JSON(http.StatusOK, updatedUser)
}

func deleteUser(c *gin.Context) {
	session := sessions.Default(c)
	admin := session.Get("admin")
	user := c.Param("name")
	if !admin.(bool) {
		processError(c, http.StatusUnauthorized, "you are not authorized to edit this user")
	}
	if err := database.DeleteUser(user); err != nil {
		processError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func getUsers(c *gin.Context) {
	users, err := database.GetAllUsers()
	if err != nil {
		processError(c, http.StatusInternalServerError, err.Error())
		return
	}
	returnedUser := []models.User{}
	for _, user := range users {
		user.Password = ""
		returnedUser = append(returnedUser, user)
	}
	c.JSON(http.StatusOK, returnedUser)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	return string(bytes), err
}
