package models

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorMessage represents an error message to be return to UI.
type ErrorMessage struct {
	Status  string
	Message string
}

// Process returns an error message to UI.
func (e *ErrorMessage) Process(c *gin.Context) {
	slog.Error(e.Message, "status", e.Status)
	c.HTML(http.StatusOK, "error", e)
	c.Abort()
}
