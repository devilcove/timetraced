package models

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorMessage struct {
	Status  string
	Message string
}

func (e *ErrorMessage) Process(c *gin.Context) {
	slog.Error(e.Message, "status", e.Status)
	c.HTML(http.StatusOK, "error", e)
	c.Abort()
}
