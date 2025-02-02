package database

import (
	"database/sql"
	"log"
)

const dbPath = "domek.db"

type DB struct {
	sql *sql.DB
}

func NewDB() (*DB, error) {
	sql, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Error opening database - %s", err)
		return nil, err
	}

	db := &DB{sql: sql}
	db.createGardenTable()

	return db, nil
}
