package main

import (
	"net/http"
	"sort"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/gin-gonic/gin"
)

func getReport(c *gin.Context) {
	reportRequest := models.ReportRequest{}
	if err := c.BindJSON(&reportRequest); err != nil {
		processError(c, http.StatusBadRequest, "could not decode request")
		return
	}
	reportData, err := database.GetReportRecords(reportRequest)
	if err != nil {
		processError(c, http.StatusInternalServerError, err.Error())
		return
	}
	sort.Slice(reportData, func(i, j int) bool {
		return reportData[i].Project < reportData[j].Project
	})
	c.JSON(http.StatusOK, reportData)

}

func report(c *gin.Context) {
	c.HTML(http.StatusOK, "report", "")
}
