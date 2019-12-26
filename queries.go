package main

import (
	"database/sql"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/teris-io/shortid"
	"os"
	"path/filepath"
)

// NewDbConnection - get a new db connection
func NewDbConnection() (*sql.DB, error) {
	logger := NewLogger()
	cs := "postgres://blink:blink@localhost:5432/blink?sslmode=disable"

	err := runMigrations(cs)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", cs)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	logger.Info("db connected successfully")

	return db, nil
}

func runMigrations(cs string) error {
	logger := NewLogger()
	logger.Info("Running migrations")
	ex, err := os.Executable()
	path := filepath.Dir(ex)
	if err != nil {
		return nil
	}
	m, err := migrate.New("file://"+path+"/migrations", cs)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	logger.Info("Migrations successfully run")
	return nil
}

// GetPath - get or insert and return a path for a provided url
func GetPath(db *sql.DB, url string) (string, error) {
	logger := NewLogger()
	var path string
	sqlString := `
		SELECT id
		FROM link
		WHERE url = $1
	`
	row := db.QueryRow(sqlString, url)
	err := row.Scan(&path)

	if err == nil {
		logger.WithFields(logrus.Fields{
			"path": path,
			"url":  url,
		}).Info("Path found for url")
		return path, nil
	}

	if err != sql.ErrNoRows {
		logger.WithFields(logrus.Fields{
			"err":  err,
			"path": path,
		}).Error("Error fetching path")
		return "", err
	}

	logger.WithFields(logrus.Fields{
		"url": url,
	}).Info("Path not found for url, generating")

	id, err := shortid.Generate()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"err": err,
			"url": url,
		}).Error("Error generating id")
		return "", err
	}

	sqlString = `
		INSERT INTO link
			(id, url)
		VALUES
			($1, $2)
	`
	_, err = db.Exec(sqlString, id, url)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"err":  err,
			"path": id,
			"url":  url,
		}).Error("Error inserting record")
		return "", nil
	}

	logger.WithFields(logrus.Fields{
		"url":  url,
		"path": path,
	}).Info("Record inserted and returned")
	return id, err
}

// GetUrl - get a url if exist for provided path
func GetUrl(db *sql.DB, path string) (string, error) {
	logger := NewLogger()
	var url string
	sqlString := `
		SELECT url
		FROM link
		WHERE id = $1
	`

	row := db.QueryRow(sqlString, path)
	err := row.Scan(&url)

	if err == nil {
		logger.WithFields(logrus.Fields{
			"url":  url,
			"path": path,
		}).Info("Url found for path")
		return url, nil
	}
	if err != sql.ErrNoRows {
		logger.WithFields(logrus.Fields{
			"err":  err,
			"path": path,
		}).Error("Error fetching url for path")
		return "", err
	}

	logger.WithFields(logrus.Fields{
		"path": path,
	}).Info("Record not found for path")
	return "", nil
}
