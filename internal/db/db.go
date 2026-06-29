package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/APIENG/apieng/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

var (
	dbPool *sql.DB
	dbLock sync.Mutex
)

// GetDB returns the singleton database connection pool
func GetDB() (*sql.DB, error) {
	if dbPool != nil {
		return dbPool, nil
	}

	dbLock.Lock()
	defer dbLock.Unlock()

	if dbPool != nil {
		return dbPool, nil
	}

	return InitializeDB()
}

// InitializeDB initializes the SQLite database and creates the necessary table.
func InitializeDB() (*sql.DB, error) {
	if dbPool != nil {
		return dbPool, nil
	}

	db, err := sql.Open("sqlite3", "./metrics.db")
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

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
		energy_consumption REAL,
		explanation TEXT
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

	createTableQuery2 := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		session_token TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME NOT NULL
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(createTableQuery1)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(createTableQuery2)
	if err != nil {
		return nil, err
	}

	dbPool = db
	return db, nil
}

// StoreMetrics stores the collected metrics in the database.
func StoreMetrics(db *sql.DB, m models.Metrics) error {
	if m.APIEndpoint == "" {
		return fmt.Errorf("API endpoint is empty")
	}

	insertQuery := `
	INSERT INTO metrics (api_endpoint, request_size, response_size, response_time, timestamp, energy_consumption, user_id, method, status, explanation)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	_, err := db.Exec(insertQuery, m.APIEndpoint, m.RequestSize, m.ResponseSize, m.ResponseTime.Milliseconds(), m.Timestamp, m.EnergyConsumption, m.UserId, m.Method, m.Status, m.Explanation)
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

// StoreSession saves a valid session token to the database
func StoreSession(db *sql.DB, userID, sessionToken string, expiresAt time.Time) error {
	insertQuery := `INSERT INTO sessions (user_id, session_token, expires_at) VALUES (?, ?, ?);`
	_, err := db.Exec(insertQuery, userID, sessionToken, expiresAt)
	return err
}

// ValidateSession checks if a session token is valid and not expired
func ValidateSession(db *sql.DB, sessionToken string) (string, error) {
	var userID string
	var expiresAt time.Time

	query := `SELECT user_id, expires_at FROM sessions WHERE session_token = ?`
	err := db.QueryRow(query, sessionToken).Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("session not found")
		}
		return "", err
	}

	if time.Now().After(expiresAt) {
		return "", fmt.Errorf("session expired")
	}

	return userID, nil
}

// InvalidateSession removes a session token
func InvalidateSession(db *sql.DB, sessionToken string) error {
	deleteQuery := `DELETE FROM sessions WHERE session_token = ?`
	_, err := db.Exec(deleteQuery, sessionToken)
	return err
}
