package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/devilcove/cookie"
	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"golang.org/x/crypto/bcrypt"
)

const SessionAge = 60 * 60 * 8 // 8 hours in seconds

func login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	user.Username = r.FormValue("username")
	user.Password = r.FormValue("password")
	if !validateUser(&user) {
		slog.Debug("validation error", "user", user.Username, "pass", user.Password)
		processError(w, http.StatusBadRequest, "invalid user")
		return
	}
	user.Password = ""
	saveCookie(user, w)
	slog.Debug("login", "user", user.Username)
	page := populatePage(user.Username)
	render(w, "content", page)
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
	slog.Error("validation failed")
	return false
}

func checkPassword(plain, hash *models.User) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash.Password), []byte(plain.Password))
	if err != nil {
		slog.Debug("bcrypt", "error", err)
	}
	return err == nil
}

func logout(w http.ResponseWriter, r *http.Request) {
	user := getRequestUser(r)
	if err := stopE(user.Username); err != nil {
		slog.Error("failed to stop tracking for user on logout", "error", err)
	}
	if err := cookie.Clear(w, cookieName, false); err != nil {
		slog.Error("clear cookie", "error", err)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func register(w http.ResponseWriter, _ *http.Request) {
	render(w, "register", models.GetPage())
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	user.Username = r.FormValue("username")
	user.Password = r.FormValue("password")
	if _, err := database.GetUser(user.Username); err == nil {
		processError(w, http.StatusBadRequest, "user exists")
		return
	}
	if user.Password == "" {
		processError(w, http.StatusBadRequest, "password cannot be blank")
		return
	}
	var err error
	user.Password, err = hashPassword(user.Password)
	if err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := database.SaveUser(&user); err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	slog.Info("new user added", "user", user.Username)
	http.Redirect(w, r, "/users/", http.StatusFound)
}

func editUser(w http.ResponseWriter, r *http.Request) {
	editor := getRequestUser(r)
	if err := r.ParseForm(); err != nil {
		processError(w, http.StatusBadRequest, "invalid data")
		return
	}
	user := models.User{
		Username: r.PathValue("name"),
		Password: r.FormValue("password"),
	}
	if user.Password == "" {
		processError(w, http.StatusBadRequest, "password cannot be blank")
		return
	}
	if r.FormValue("admin") != "" {
		user.IsAdmin = true
	}
	slog.Debug("edit user", "new", user)
	if user.Username != editor.Username && !editor.IsAdmin {
		processError(w, http.StatusUnauthorized, "you are not authorized to edit this user")
		return
	}
	updatedUser, err := database.GetUser(user.Username)
	if err != nil {
		processError(w, http.StatusBadRequest, "user does not exist"+user.Username)
		return
	}
	updatedUser.Password, err = hashPassword(user.Password)
	if err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updatedUser.IsAdmin = user.IsAdmin
	if !editor.IsAdmin {
		updatedUser.IsAdmin = false
	}
	updatedUser.Updated = time.Now()
	slog.Debug("updating user", "old", user, "new", updatedUser)
	if err := database.SaveUser(&updatedUser); err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	slog.Info("user updated", "user", updatedUser.Username)
	http.Redirect(w, r, "/users/", http.StatusFound)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	editor := getRequestUser(r)
	user := r.PathValue("name")
	if !editor.IsAdmin {
		processError(w, http.StatusUnauthorized, "you are not authorized to delete this user")
		return
	}
	if _, err := database.GetUser(user); err != nil {
		processError(w, http.StatusBadRequest, "user does not exist")
		return
	}
	if err := database.DeleteUser(user); err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	slog.Info("deleted", "user", user)
	getUsers(w, r)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	editor := getRequestUser(r)
	if !editor.IsAdmin {
		getCurrentUser(w, editor.Username)
		return
	}
	users, err := database.GetAllUsers()
	if err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	returnedUser := []models.User{}
	for _, user := range users {
		user.Password = ""
		returnedUser = append(returnedUser, user)
	}
	render(w, "user", returnedUser)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	editor := getRequestUser(r)
	editUser := r.PathValue("name")
	if !editor.IsAdmin && editUser != editor.Username {
		processError(w, http.StatusBadRequest, "non-admin cannot edit other users")
		return
	}
	user, err := database.GetUser(r.PathValue("name"))
	if err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	edit := models.Editor{User: user, AsAdmin: editor.IsAdmin}
	render(w, "editUser", edit)
}

func getCurrentUser(w http.ResponseWriter, name string) {
	user, err := database.GetUser(name)
	if err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	editor := models.Editor{User: user, AsAdmin: false}
	render(w, "editUser", editor)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	return string(bytes), err
}
