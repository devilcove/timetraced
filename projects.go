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

func displayProjectForm(w http.ResponseWriter, _ *http.Request) {
	render(w, "addProject", nil)
}

func showProjects(w http.ResponseWriter, r *http.Request) {
	user := getRequestUser(r)
	render(w, "showProjects", populatePage(user.Username))
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

func start(w http.ResponseWriter, r *http.Request) {
	proj := r.PathValue("name")
	user := getRequestUser(r)
	project, err := database.GetProject(proj)
	if err != nil {
		processError(w, http.StatusBadRequest, "error reading project "+err.Error())
		return
	}
	if !project.Active {
		processError(w, http.StatusBadRequest, "project is not active")
		return
	}
	if models.IsTrackingActive(user.Username) {
		if err := stopE(user.Username); err != nil {
			processError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	record := models.Record{
		ID:      uuid.New(),
		Project: proj,
		User:    user.Username,
		Start:   time.Now(),
	}
	if err := database.SaveRecord(&record); err != nil {
		processError(w, http.StatusInternalServerError, "failed to save record "+err.Error())
		return
	}
	models.TrackingActive(user.Username, project)
	slog.Info("tracking started", "project", project.Name)
	render(w, "content", populatePage(user.Username))
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
	user := getRequestUser(r)
	if err := stopE(user.Username); err != nil {
		processError(w, http.StatusInternalServerError, err.Error())
		return
	}
	render(w, "content", populatePage(user.Username))
}
