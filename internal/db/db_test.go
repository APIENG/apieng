package db

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func createTestDB(t *testing.T) *sql.DB {
	dbFile := fmt.Sprintf("test_metrics_%d.db", time.Now().UnixNano())
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
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

	db.Exec(createTableQuery)
	db.Exec(createTableQuery1)
	db.Exec(createTableQuery2)

	t.Cleanup(func() {
		db.Close()
		os.Remove(dbFile)
	})

	return db
}

func TestInitializeDB(t *testing.T) {
	db := createTestDB(t)

	// Verify tables exist by querying
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='users'")
	if err != nil {
		t.Fatalf("Failed to query tables: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("users table was not created")
	}
}

func TestStoreAndValidateSession(t *testing.T) {
	db := createTestDB(t)

	userID := "test_user_123"
	sessionToken := "session_token_456"
	expiresAt := time.Now().Add(24 * time.Hour)

	// Test storing a session
	err := StoreSession(db, userID, sessionToken, expiresAt)
	if err != nil {
		t.Fatalf("Failed to store session: %v", err)
	}

	// Test validating the session
	validatedUserID, err := ValidateSession(db, sessionToken)
	if err != nil {
		t.Fatalf("Failed to validate session: %v", err)
	}

	if validatedUserID != userID {
		t.Fatalf("Expected user ID %s, got %s", userID, validatedUserID)
	}
}

func TestInvalidateSession(t *testing.T) {
	db := createTestDB(t)

	userID := "test_user_123"
	sessionToken := "session_token_456"
	expiresAt := time.Now().Add(24 * time.Hour)

	err := StoreSession(db, userID, sessionToken, expiresAt)
	if err != nil {
		t.Fatalf("Failed to store session: %v", err)
	}

	// Invalidate the session
	err = InvalidateSession(db, sessionToken)
	if err != nil {
		t.Fatalf("Failed to invalidate session: %v", err)
	}

	// Try to validate the invalidated session
	_, err = ValidateSession(db, sessionToken)
	if err == nil {
		t.Fatal("Expected error when validating invalidated session")
	}
}
