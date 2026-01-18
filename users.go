package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

const SessionAge = 60 * 60 * 8 // 8 hours in seconds

func login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := r.ParseForm(); err != nil {
		processError(w, http.StatusBadRequest, "invalid user form")
		return
	}
	user.Username = r.FormValue("username")
	user.Password = r.FormValue("password")
	if !validateUser(&user) {
		slog.Debug("validation error", "user", user.Username, "pass", user.Password)
		processError(w, http.StatusBadRequest, "invalid user")
		return
	}
	session := sessions.NewSession(store, "devilcove-time")
	session.Values["user"] = user.Username
	session.Values["loggedIn"] = true
	session.Values["admin"] = user.IsAdmin
	session.Options = &sessions.Options{
		MaxAge:   SessionAge,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	if err := session.Save(r, w); err != nil {
		slog.Error("session save", "error", err)
	}
	user.Password = ""
	slog.Debug("login", "user", user.Username)
	page := populatePage(user.Username)
	page.NeedsLogin = false
	projects, err := database.GetAllProjects()
	if err != nil {
		slog.Error(err.Error())
	} else {
		for _, project := range projects {
			page.Projects = append(page.Projects, project.Name)
		}
	}
	_ = templates.ExecuteTemplate(w, "content", page)
}

func validateUser(visitor *models.User) bool {
	user, err := database.GetUser(visitor.Username)
	if err != nil {
		slog.Error("no such user", "user", visitor.Username, "error", err)
		return false
	}
	// fmt.Println(visitor.Username, user.Username)
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
	session := sessionData(r)
	if session != nil {
		if err := stopE(session.User); err != nil {
			slog.Error("failed to stop tracking for user on logout", "error", err)
		}
		session.Session.Options.MaxAge = -1
		_ = session.Session.Save(r, w)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func register(w http.ResponseWriter, _ *http.Request) {
	page := models.GetPage()
	_ = templates.ExecuteTemplate(w, "register", page)
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := r.ParseForm(); err != nil {
		processError(w, http.StatusBadRequest, "invalid user")
		return
	}
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
	session := sessionData(r)
	if session == nil {
		displayMain(w, r)
		return
	}
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
	visitor := session.User
	if user.Username != visitor && !session.Admin {
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
	if !session.Admin {
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
	session := sessionData(r)
	if session == nil {
		displayMain(w, r)
		return
	}
	user := r.PathValue("name")
	if !session.Admin {
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
	session := sessionData(r)
	if session == nil {
		processError(w, http.StatusBadRequest, "no session")
		return
	}
	if !session.Admin {
		getCurrentUser(w, r)
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
	slog.Info("getusers", "users", returnedUser, "session", session)
	_ = templates.ExecuteTemplate(w, "user", returnedUser)
	// c.HTML(http.StatusOK, "user", returnedUser)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	session := sessionData(r)
	if session == nil {
		processError(w, http.StatusBadRequest, "no session")
		return
	}
	editUser := r.PathValue("name")
	if !session.Admin && editUser != session.User {
		processError(w, http.StatusBadRequest, "non-admin cannot edit other users")
		return
	}
	user, err := database.GetUser(editUser)
	if err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	editor := models.Editor{User: user, AsAdmin: session.Admin}
	_ = templates.ExecuteTemplate(w, "editUser", editor)
}

func getCurrentUser(w http.ResponseWriter, r *http.Request) {
	session := sessionData(r)
	if session == nil {
		processError(w, http.StatusBadRequest, "session data")
		return
	}
	user, err := database.GetUser(session.User)
	if err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	editor := models.Editor{User: user, AsAdmin: session.Admin}
	_ = templates.ExecuteTemplate(w, "editUser", editor)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	return string(bytes), err
}
