/*
Copyright © 2022 Matthew R Kasun <mkasun@nusak.ca>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// @title TimeTracker
// @version v0.1.0
// @description time tracking application
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host time.nusak.ca
package main

import (
	"log"
	"os"
	"path/filepath"

	"log/slog"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/joho/godotenv"
	sloggin "github.com/samber/slog-gin"
)

func main() {
	logger := setLogging()
	if err := godotenv.Load(); err != nil {
		slog.Error("read environment", "error", err)
	}
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}
	database.InitializeDatabase()
	defer database.Close()
	checkDefaultUser()
	project := database.GetActiveProject()
	if project != nil {
		models.TrackingActive(*project)
	} else {
		models.TrackingInactive()
	}
	router := setupRouter()
	router.Use(sloggin.New(logger))
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func setLogging() *slog.Logger {
	logLevel := &slog.LevelVar{}
	replace := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			source, _ := a.Value.Any().(*slog.Source)
			if source != nil {
				source.File = filepath.Base(source.File)
				source.Function = filepath.Base(source.Function)
			}
		}
		return a
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, ReplaceAttr: replace, Level: logLevel}))
	//logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: logLevel}))
	slog.SetDefault(logger)
	if os.Getenv("DEBUG") == "true" {
		logLevel.Set(slog.LevelDebug)
	}
	return logger
}
