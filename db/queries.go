package db

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
	"github.com/teris-io/shortid"
)

// NewDbConnection - get a new db connection
func NewDbConnection() (*sql.DB, error) {
	cs := "postgres://blink:blink@localhost:5432/blink?sslmode=disable"

	err := runMigrations(cs)
	if err != nil {
		return nil, err
	}
	fmt.Println("migrations run successfully")

	db, err := sql.Open("postgres", cs)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	fmt.Println("db connected successfully")

	return db, nil
}

func runMigrations(cs string) error {
	fmt.Println("Running migrations")
	m, err := migrate.New("file://db/migrations", cs)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	fmt.Println("Migrations successfully run")
	return nil
}

// GetPath - get or insert and return a path for a provided url
func GetPath(db *sql.DB, url string) (string, error) {
	var path string
	sqlString := `
		SELECT id
		FROM link
		WHERE url = $1
	`
	row := db.QueryRow(sqlString, url)
	err := row.Scan(&path)

	if err == nil {
		return path, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}

	id, err := shortid.Generate()
	sqlString = `
		INSERT INTO link
			(id, url)
		VALUES
			($1, $2)
	`
	_, err = db.Exec(sqlString, shortid.MustGenerate(), url)
	if err != nil {
		return "", nil
	}
	return id, err
}

// GetUrl - get a url if exist for provided path
func GetUrl(db *sql.DB, path string) (string, error) {
	var url string
	sqlString := `
		SELECT url
		FROM link
		WHERE id = $1
	`
	row := db.QueryRow(sqlString, path)
	err := row.Scan(&url)

	if err == nil {
		return url, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}
	return "", nil
}
