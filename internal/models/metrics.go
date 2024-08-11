package models

import (
	"time"
	"golang.org/x/crypto/bcrypt"
)


type Users struct {
	Email    string
	Password string
}

// Metrics represents the data structure for storing API metrics.
type Metrics struct {
	APIEndpoint       string
	RequestSize       int
	ResponseSize      int
	ResponseTime      time.Duration
	Timestamp         time.Time
	EnergyConsumption float64
}

func (u *Users) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}


func (u *Users) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}