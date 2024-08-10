package db

import (
	"database/sql"
	"fmt"

	"github.com/Taiwrash/apieng/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// InitializeDB initializes the SQLite database and creates the necessary table.
func InitializeDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./metrics.db")
	if err != nil {
		return nil, err
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		api_endpoint TEXT,
		request_size INTEGER,
		response_size INTEGER,
		response_time INTEGER,
		timestamp DATETIME,
		energy_consumption REAL
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// StoreMetrics stores the collected metrics in the database.
func StoreMetrics(db *sql.DB, m models.Metrics) error {
	if m.APIEndpoint == "" {
		return fmt.Errorf("API endpoint is empty")
	}

	insertQuery := `
	INSERT INTO metrics (api_endpoint, request_size, response_size, response_time, timestamp, energy_consumption)
	VALUES (?, ?, ?, ?, ?, ?);`

	_, err := db.Exec(insertQuery, m.APIEndpoint, m.RequestSize, m.ResponseSize, m.ResponseTime.Milliseconds(), m.Timestamp, m.EnergyConsumption)
	if err != nil {
		return err
	}

	return nil
}
