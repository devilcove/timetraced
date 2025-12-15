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
	"os"
	"time"

	"github.com/devilcove/timetraced/database"
	"github.com/devilcove/timetraced/models"
	"github.com/mattkasun/tools/logging"
)

func main() {
	logger := logging.TextLogger(logging.TruncateSource(), logging.TimeFormat(time.DateTime))
	//logger := logging.TextLogger(logging.TruncateSource(), logging.TimeFormat(time.DateTime), logging.Level(slog.LevelDebug))
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	if err := database.InitializeDatabase(); err != nil {
		logger.Error("database init", "err", err)
		os.Exit(1)
	}
	defer database.Close()
	checkDefaultUser()
	users, err := database.GetAllUsers()
	if err != nil {
		logger.Error("get users", "err", err)
		os.Exit(1) //nolint:gocritic
	}
	for _, user := range users {
		project := database.GetActiveProject(user.Username)
		if project != nil {
			models.TrackingActive(user.Username, *project)
		} else {
			models.TrackingInactive(user.Username)
		}
	}
	router := setupRouter(logger.Logger)
	router.Run(":" + port)
}
