package services

import (
	"github.com/google/uuid"
)

func GenerateSessionID() string {
	// Generate a new UUID for the session ID
	sessionID := uuid.New().String()
	return sessionID
}
