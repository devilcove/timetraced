package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const SessionAge = 60 * 60 * 8 // 8 hours in seconds

func displayLogin(c *gin.Context) {
	page := models.GetPage()
	page.Page = "login"
	c.HTML(http.StatusOK, "loginFull", page)
}

func login(c *gin.Context) {
	session := sessions.Default(c)
	var user models.User
	if err := c.Bind(&user); err != nil {
		processError(c, "bad request", "invalid user")
		slog.Error("bind err", "error", err)
		return
	}
	slog.Debug("login by", "user", user)
	if !validateUser(&user) {
		session.Clear()
		session.Save()
		processError(c, "bad request", "invalid user")
		slog.Warn("validation error", "user", user.Username)
		return
	}
	session.Set("loggedin", true)
	session.Set("user", user.Username)
	session.Set("admin", user.IsAdmin)
	session.Options(sessions.Options{MaxAge: SessionAge, Secure: false, SameSite: http.SameSiteLaxMode})
	session.Save()
	user.Password = ""
	slog.Info("login", "user", user.Username)
	location := url.URL{Path: "/"}
	c.Redirect(http.StatusFound, location.RequestURI())
}

func validateUser(visitor *models.User) bool {
	user, err := database.GetUser(visitor.Username)
	if err != nil {
		slog.Error("no such user", "user", visitor.Username, "error", err)
		return false
	}
	fmt.Println(visitor.Username, user.Username)
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
	location := url.URL{Path: "/"}
	c.Redirect(http.StatusFound, location.RequestURI())
}

func register(c *gin.Context) {
	c.HTML(http.StatusOK, "registerFull", models.GetPage())
}

func regUser(c *gin.Context) {
	var user models.User
	var err error
	if err := c.Bind(&user); err != nil {
		processError(c, "bad request", err.Error())
		return
	}
	if _, err := database.GetUser(user.Username); err == nil {
		processError(c, "bad request", "user exists")
		return
	}
	if user.Password == "" {
		processError(c, "bad request", "password cannot be blank")
		return
	}
	user.Password, err = hashPassword(user.Password)
	if err != nil {
		processError(c, "ServerError", err.Error())
		return
	}
	if err := database.SaveUser(&user); err != nil {
		processError(c, "ServerError", err.Error())
		return
	}
	slog.Info("new user added", "user", user.Username)
	location := url.URL{Path: "/"}
	c.Redirect(http.StatusFound, location.RequestURI())
}

func addUser(c *gin.Context) {
	var user models.User
	var err error
	session := sessions.Default(c)
	admin := session.Get("admin")
	if !admin.(bool) {
		processError(c, "unauthorized", "only admins can create new users")
	}
	if err := c.BindJSON(&user); err != nil {
		processError(c, "bad request", "could not decode request into json")
		return
	}
	if _, err := database.GetUser(user.Username); err == nil {
		processError(c, "bad request", "user exists")
		return
	}
	if user.Username == "" || user.Password == "" {
		processError(c, "bad request", "username or password cannot be blank")
		return
	}
	user.Password, err = hashPassword(user.Password)
	if err != nil {
		processError(c, "ServerError", err.Error())
		return
	}
	if err := database.SaveUser(&user); err != nil {
		processError(c, "ServerError", err.Error())
		return
	}
	slog.Info("new user added", "user", user.Username)
	c.JSON(http.StatusNoContent, nil)
}

func editUser(c *gin.Context) {
	var user struct {
		Password string
		Verify   string
	}
	var err error
	username := c.Param("name")
	session := sessions.Default(c)
	admin := session.Get("admin")
	visitor := session.Get("user").(string)
	if err := c.Bind(&user); err != nil {
		processError(c, "bad request", "could not decode request into json")
		return
	}
	if username != visitor && !admin.(bool) {
		processError(c, "unauthorized", "you are not authorized to edit this user")
		return
	}
	updatedUser, err := database.GetUser(username)
	if err != nil {
		processError(c, "bad request", "user does not exist")
		return
	}
	updatedUser.Password, err = hashPassword(user.Password)
	if err != nil {
		processError(c, "ServerError", err.Error())
		return
	}
	if !admin.(bool) {
		updatedUser.IsAdmin = false
	}
	updatedUser.Updated = time.Now()
	if err := database.SaveUser(&updatedUser); err != nil {
		processError(c, "ServerError", err.Error())
		return
	}
	slog.Info("user updated", "user", updatedUser.Username)
	location := url.URL{Path: "/"}
	c.Redirect(http.StatusFound, location.RequestURI())
}

func deleteUser(c *gin.Context) {
	session := sessions.Default(c)
	admin := session.Get("admin")
	user := c.Param("name")
	if !admin.(bool) {
		processError(c, "unauthorized", "you are not authorized to delete this user")
		return
	}
	if _, err := database.GetUser(user); err != nil {
		processError(c, "bad request", "user does not exist")
		return
	}
	if err := database.DeleteUser(user); err != nil {
		processError(c, "ServerError", err.Error())
		return
	}
	slog.Info("deleted", "user", user)
	c.JSON(http.StatusNoContent, nil)
}

func getUsers(c *gin.Context) {
	session := sessions.Default(c)
	admin := session.Get("admin").(bool)
	if !admin {
		getCurrentUser(c)
		return
	}
	users, err := database.GetAllUsers()
	if err != nil {
		processError(c, "ServerError", err.Error())
		return
	}
	returnedUser := []models.User{}
	for _, user := range users {
		user.Password = ""
		returnedUser = append(returnedUser, user)
	}
	c.HTML(http.StatusOK, "user", returnedUser)
}

func getUser(c *gin.Context) {
	session := sessions.Default(c)
	editUser := c.Param("name")
	if !session.Get("admin").(bool) && editUser != session.Get("user").(string) {
		processError(c, "bad request", "non-admin cannot edit other users")
		return
	}
	user, err := database.GetUser(editUser)
	if err != nil {
		processError(c, "ServerError", err.Error())
		return
	}
	c.HTML(http.StatusOK, "editUser", user)
}

func getCurrentUser(c *gin.Context) {
	session := sessions.Default(c)
	visitor := session.Get("user").(string)
	user, err := database.GetUser(visitor)
	if err != nil {
		processError(c, "ServerError", err.Error())
		return
	}
	c.HTML(http.StatusOK, "editUser", user)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	return string(bytes), err
}
