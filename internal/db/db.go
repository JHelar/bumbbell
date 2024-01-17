package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	DB *sql.DB
}

const file string = "database.db"

const create string = `
  CREATE TABLE IF NOT EXISTS users (
  id TEXT NOT NULL PRIMARY KEY
  );`

func NewDB() (*DB, error) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(create); err != nil {
		return nil, err
	}

	return &DB{
		DB: db,
	}, nil
}
