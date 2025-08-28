// Package database manages persistent storage for timetraced using bbolt.
// It initializes and closes the database, and provides functions to create,
// retrieve, update, and delete projects, users, and records. The package
// also includes queries for common application needs such as active projects,
// daily records, and report generation.
package database
