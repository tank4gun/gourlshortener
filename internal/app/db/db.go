package db

import (
	"log"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// RunMigrations - apply all needed migrations to the db
func RunMigrations(dbDSN string) error {
	if dbDSN == "" {
		return nil
	}
	m, err := migrate.New(
		"file://internal/app/db/migrations",
		dbDSN)
	if err != nil {
		log.Fatal(err)
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
		return err
	}
	return nil
}
