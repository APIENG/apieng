package models

import (
	"testing"
)

func TestSetAndCheckPassword(t *testing.T) {
	user := &Users{
		Email: "test@example.com",
		Iid:   "user_123",
	}

	password := "securePassword123"

	// Test SetPassword
	err := user.SetPassword(password)
	if err != nil {
		t.Fatalf("Failed to set password: %v", err)
	}

	// Password should be hashed, not the original
	if user.Password == password {
		t.Fatal("Password was not hashed")
	}

	// Test CheckPassword with correct password
	if !user.CheckPassword(password) {
		t.Fatal("CheckPassword failed with correct password")
	}

	// Test CheckPassword with wrong password
	if user.CheckPassword("wrongPassword") {
		t.Fatal("CheckPassword succeeded with wrong password")
	}
}
