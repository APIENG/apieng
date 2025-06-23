package models

import (
	"time"
)

// Metrics represents the data structure for storing API metrics.
type Metrics struct {
	ID                int           `json:"id" db:"id"`
	APIEndpoint       string        `json:"api_endpoint" db:"api_endpoint"`
	UserId            string        `json:"user_id" db:"user_id"`
	Method            string        `json:"method" db:"method"`
	Status            int           `json:"status" db:"status"`
	RequestSize       int           `json:"request_size" db:"request_size"`
	ResponseSize      int           `json:"response_size" db:"response_size"`
	ResponseTime      time.Duration `json:"response_time" db:"response_time"`
	Timestamp         time.Time     `json:"timestamp" db:"timestamp"`
	EnergyConsumption float64       `json:"energy_consumption" db:"energy_consumption"`
	Explanation       string        `json:"explanation" db:"explanation"`
}
