package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/google/uuid"
)

// func getProjects(w http.ResponseWriter, r *http.Request) {
// 	projects, err := database.GetAllProjects()
// 	if err != nil {
// 		processError(w, http.StatusInternalServerError, err.Error())
// 		return
// 	}
// 	c.JSON(http.StatusOK, projects)
// }

func displayProjectForm(w http.ResponseWriter, _ *http.Request) {
	_ = templates.ExecuteTemplate(w, "addProject", nil)
}

func addProject(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		processError(w, http.StatusBadRequest, "invalid data sent")
	}
	project := models.Project{
		Name: r.FormValue("name"),
	}
	if regexp.MustCompile(`\s+`).MatchString(project.Name) || project.Name == "" {
		processError(w, http.StatusBadRequest, "invalid project name")
		return
	}
	existing, err := database.GetProject(project.Name)
	if err != nil && err.Error() != "no such project" {
		slog.Error("add project", "error", err)
		processError(w, http.StatusInternalServerError, "database error")
		return
	}
	if existing.Name == project.Name {
		processError(w, http.StatusBadRequest, "project exists")
		return
	}
	project.ID = uuid.New()
	project.Active = true
	project.Updated = time.Now()
	if err := database.SaveProject(&project); err != nil {
		processError(w, http.StatusInternalServerError, "error saving project "+err.Error())
		return
	}
	displayMain(w, r)
}

// func getProject(w http.ResponseWriter, r *http.Request) {
// 	p := r.PathValue("name")
// 	project, err := database.GetProject(p)
// 	if err != nil {
// 		processError(w, http.StatusBadRequest, "could not retrieve project "+err.Error())
// 		return
// 	}

// 	c.JSON(http.StatusOK, project)
// }

func start(w http.ResponseWriter, r *http.Request) {
	proj := r.PathValue("name")
	session := sessionData(r)
	if session == nil {
		displayMain(w, r)
		return
	}
	user := session.User
	project, err := database.GetProject(proj)
	if err != nil {
		processError(w, http.StatusInternalServerError, "error reading project "+err.Error())
		return
	}
	if !project.Active {
		processError(w, http.StatusBadRequest, "project is not active")
		return
	}
	if models.IsTrackingActive(user) {
		if err := stopE(user); err != nil {
			processError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	record := models.Record{
		ID:      uuid.New(),
		Project: proj,
		User:    user,
		Start:   time.Now(),
	}
	if err := database.SaveRecord(&record); err != nil {
		processError(w, http.StatusInternalServerError, "failed to save record "+err.Error())
		return
	}
	models.TrackingActive(user, project)
	slog.Info("tracking started", "project", project.Name)
	displayMain(w, r)
}

func stopE(user string) error {
	records, err := database.GetAllRecordsForUser(user)
	if err != nil {
		return fmt.Errorf("failed to retrieve records %w", err)
	}
	for _, record := range records {
		if record.End.IsZero() {
			record.End = time.Now()
			if err := database.SaveRecord(&record); err != nil {
				slog.Error("failed to save updated record", "error", err)
			}
		}
	}
	slog.Info("tracking stopped", "project", models.Tracked(user))
	models.TrackingInactive(user)
	return nil
}

func stop(w http.ResponseWriter, r *http.Request) {
	session := sessionData(r)
	user := session.User
	if err := stopE(user); err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	displayMain(w, r)
}
