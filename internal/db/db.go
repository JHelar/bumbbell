package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const file string = "db/database.db"

func NewDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}

	return db, nil
}
