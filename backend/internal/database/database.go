package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"github.com/mattpotok/domek/backend/internal/common"
)

const DB_FILE_NAME = "domek.db"

type DB struct {
	sql *sql.DB
}

func NewDB() (*DB, error) {
	db_file_path := filepath.Join(common.GetServiceDirPath(), DB_FILE_NAME)

	_, err := os.Stat(db_file_path)
	if os.IsNotExist(err) {
		file, err := os.Create(db_file_path)
		if err != nil {
			return nil, err
		}

		if err = file.Close(); err != nil {
			return nil, err
		}
	}

	sql, err := sql.Open("sqlite3", db_file_path)
	if err != nil {
		log.Fatalf("Error opening database - %s", err)
		return nil, err
	}

	db := &DB{sql: sql}
	db.createGardenTable()

	return db, nil
}
