package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const SessionAge = 60 * 60 * 8 // 8 hours in seconds

func login(c *gin.Context) {
	session := sessions.Default(c)
	var user models.User
	if err := c.BindJSON(&user); err != nil {
		processError(c, http.StatusBadRequest, "invalid user")
		slog.Error("bind err", "error", err)
		return
	}
	slog.Debug("login by", "user", user)
	if !validateUser(&user) {
		session.Clear()
		session.Save()
		processError(c, http.StatusBadRequest, "invalid user")
		slog.Warn("validation error", "user", user.Username)
		return
	}
	session.Set("loggedin", true)
	session.Set("user", user.Username)
	session.Set("admin", user.IsAdmin)
	session.Options(sessions.Options{MaxAge: SessionAge, Secure: true, SameSite: http.SameSiteLaxMode})
	session.Save()
	user.Password = ""
	slog.Info("login", "user", user.Username)
	c.JSON(http.StatusOK, user)
}

func validateUser(visitor *models.User) bool {
	user, err := database.GetUser(visitor.Username)
	if err != nil {
		slog.Error("no such user", "user", visitor.Username, "error", err)
		return false
	}
	if visitor.Username == user.Username && checkPassword(visitor, &user) {
		visitor.IsAdmin = user.IsAdmin
		return true
	}
	return false
}

func checkPassword(plain, hash *models.User) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash.Password), []byte(plain.Password))
	if err != nil {
		slog.Debug("bcrypt", "error", err)
	}
	return err == nil
}

func logout(c *gin.Context) {
	session := sessions.Default(c)
	slog.Info("logout", "user", session.Get("user"))
	//delete cookie
	session.Clear()
	session.Save()
	c.JSON(http.StatusNoContent, nil)
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
	if _, err := database.GetUser(user.Username); err == nil {
		processError(c, http.StatusBadRequest, "user exists")
		return
	}
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
	slog.Info("new user added", "user", user.Username)
	c.JSON(http.StatusNoContent, nil)
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
		return
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
	if err := database.SaveUser(&updatedUser); err != nil {
		processError(c, http.StatusInternalServerError, err.Error())
		return
	}
	updatedUser.Password = ""
	slog.Info("user updated", "user", updatedUser.Username)
	c.JSON(http.StatusOK, updatedUser)
}

func deleteUser(c *gin.Context) {
	session := sessions.Default(c)
	admin := session.Get("admin")
	user := c.Param("name")
	if !admin.(bool) {
		processError(c, http.StatusUnauthorized, "you are not authorized to delete this user")
		return
	}
	if _, err := database.GetUser(user); err != nil {
		processError(c, http.StatusBadRequest, "user does not exist")
		return
	}
	if err := database.DeleteUser(user); err != nil {
		processError(c, http.StatusInternalServerError, err.Error())
		return
	}
	slog.Info("deleted", "user", user)
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
