package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/devilcove/timetraced/models"
	"github.com/gin-gonic/gin"
	"github.com/kr/pretty"
)

func config(c *gin.Context) {
	page := models.GetPage()
	pretty.Println(page)
	c.HTML(http.StatusOK, "config", page)
}

func setConfig(c *gin.Context) {
	config := models.Config{}
	if err := c.Bind(&config); err != nil {
		log.Println("failed to read config", err)
	}
	models.SetTheme(config.Theme)
	models.SetFont(config.Font)
	location := url.URL{Path: "/"}
	c.Redirect(http.StatusFound, location.RequestURI())
}
