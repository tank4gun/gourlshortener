package db

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func CreateDb(dbDSN string) (*sql.DB, error) {
	database, err := sql.Open("pgx", dbDSN)
	if err != nil {
		return nil, err
	}
	//defer database.Close()
	return database, nil
}
