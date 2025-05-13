package db

import (
	"database/sql"
	"fmt"

	"github.com/APIENG/apieng/internal/models"

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
		user_id TEXT,
		request_size INTEGER,
		response_size INTEGER,
		response_time INTEGER,
		method TEXT DEFAULT 'GET',
		status INTEGER DEFAULT 200,
		timestamp DATETIME,
		energy_consumption REAL
	);`

	createTableQuery1 := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		iid TEXT,
		email TEXT,
		firstname TEXT,
		lastname TEXT,
		password TEXT,
		apikey TEXT
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(createTableQuery1)
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
	INSERT INTO metrics (api_endpoint, request_size, response_size, response_time, timestamp, energy_consumption, user_id, method, status)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);`

	_, err := db.Exec(insertQuery, m.APIEndpoint, m.RequestSize, m.ResponseSize, m.ResponseTime.Milliseconds(), m.Timestamp, m.EnergyConsumption, m.UserId, m.Method, m.Status)
	if err != nil {
		return err
	}

	return nil
}

// StoreUsers stores the collected users in the database.
func StoreUsers(db *sql.DB, u models.Users) error {

	// log.Println(u.Iid, u.Email, u.FirstName, u.LastName, u.Password)

	insertQuery := `
	INSERT INTO users (iid, email, firstname, lastname, password)
	VALUES (?, ?, ?, ?, ?);`

	_, err := db.Exec(insertQuery, u.Iid, u.Email, u.FirstName, u.LastName, u.Password)
	if err != nil {
		return err
	}

	return nil
}

func UpdateUser(db *sql.DB, iid string, apikey string) error {

	insertQuery := `Update users set apikey = ? where iid = ?;`

	_, err := db.Exec(insertQuery, apikey, iid)
	if err != nil {
		return err
	}

	return nil
}
