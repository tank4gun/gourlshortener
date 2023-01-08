package db

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
)

func RunMigrations(dbDSN string) error {
	if dbDSN == "" {
		return nil
	}
	m, err := migrate.New(
		"file://internal/app/db/migrations",
		fmt.Sprintf("%s?sslmode=disable", dbDSN))
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

func CreateDB(dbDSN string) (*sql.DB, error) {
	database, err := sql.Open("pgx", dbDSN)
	if err != nil {
		return nil, err
	}
	//defer database.Close()
	return database, nil
}
