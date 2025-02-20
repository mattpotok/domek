package database

import (
	"log"
)

type Vegetable struct {
	Id       int
	HasFruit bool
	Name     string
	Quantity int
	Variety  string
	Year     int
	Yield    int
}

func (db *DB) createGardenTable() error {
	gardenTable := `
		CREATE TABLE IF NOT EXISTS garden (
			id     	  INTEGER PRIMARY KEY AUTOINCREMENT,
			hasFruit  BOOLEAN NOT NULL CHECK(hasFruit in (0, 1)),
			name      TEXT    NOT NULL,
			quantity  INTEGER NOT NULL,
			variety   TEXT    NOT NULL,
			year      INTEGER NOT NULL,
			yield     INTEGER NOT NULL
		);`

	_, err := db.sql.Exec(gardenTable)
	if err != nil {
		log.Fatalf("Error creating garden table - %s", err)
	}

	return nil
}

func (db *DB) InsertVegetable(veg *Vegetable) error {
	insertVegetable := `
		INSERT INTO garden (hasFruit, name, quantity, variety, year, yield)
		VALUES (?, ?, ?, ?, ?, ?);`

	_, err := db.sql.Exec(insertVegetable, veg.HasFruit, veg.Name, veg.Quantity, veg.Variety, veg.Year, veg.Yield)
	if err != nil {
		log.Fatalf("Error inserting vegetable - %s", err)
	}

	return nil
}

func (db *DB) GetVegetables() ([]Vegetable, error) {
	query := `SELECT * FROM garden;`
	rows, err := db.sql.Query(query)
	if err != nil {
		log.Fatalf("Error getting vegetables - %s", err)
	}

	var veges []Vegetable

	for rows.Next() {
		var veg Vegetable
		err = rows.Scan(&veg.Id, &veg.HasFruit, &veg.Name, &veg.Quantity, &veg.Variety, &veg.Year, &veg.Yield)
		if err != nil {
			log.Fatalf("Error scanning vegetable - %s", err)
		}

		veges = append(veges, veg)
	}

	return veges, nil
}

func (db *DB) GetVegetablesByYear(year int) ([]Vegetable, error) {
	query := `SELECT id, hasFruit, name, quantity, variety, year, yield FROM garden WHERE year=?;`
	rows, err := db.sql.Query(query, year)
	if err != nil {
		log.Fatalf("Error getting vegetables - %s", err)
	}

	var vegetables []Vegetable
	for rows.Next() {
		var veg Vegetable
		err = rows.Scan(&veg.Id, &veg.HasFruit, &veg.Name, &veg.Quantity, &veg.Variety, &veg.Year, &veg.Yield)
		if err != nil {
			log.Fatalf("Error scanning vegetable - %s", err)
		}

		vegetables = append(vegetables, veg)
	}

	return vegetables, nil
}

func (db *DB) UpdateVegetable(veg Vegetable) error {
	query := `
		UPDATE garden
		SET hasFruit = ?, name = ?, quantity = ?, variety = ?, year = ?, yield = ?
		WHERE id = ?;`
	if _, err := db.sql.Exec(query, veg.HasFruit, veg.Name, veg.Quantity, veg.Variety, veg.Year, veg.Yield, veg.Id); err != nil {
		return err
	}

	// TODO check the number of rows affected

	return nil
}
