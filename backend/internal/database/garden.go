package database

import "log"

type Vegetable struct {
	Id       int
	HasFruit int
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
	// TODO sort the vegetables by name and then by variety
	getVegetables := `SELECT * FROM garden;`

	rows, err := db.sql.Query(getVegetables)
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
