package models

import (
	"time"
)

// Metrics represents the data structure for storing API metrics.
type Metrics struct {
	APIEndpoint       string
	RequestSize       int
	ResponseSize      int
	ResponseTime      time.Duration
	Timestamp         time.Time
	EnergyConsumption float64
}
