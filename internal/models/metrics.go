package models

import (
	"time"
)

// Metrics represents the data structure for storing API metrics.
type Metrics struct {
	APIEndpoint       string
	UserId            string
	Method            string
	Status            int
	RequestSize       int
	ResponseSize      int
	ResponseTime      time.Duration
	Timestamp         time.Time
	EnergyConsumption float64
	Explanation       string
}
