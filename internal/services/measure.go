package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/APIENG/apieng/internal/models"
)

// MeasureAPI collects metrics for a given API endpoint and estimates energy consumption.
func MeasureAPI(endpoint string, user string) models.Metrics {
	start := time.Now()
	resp, err := http.Get(endpoint)
	if err != nil {
		fmt.Println("Error making request:", err)
		return models.Metrics{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return models.Metrics{}
	}

	// For demonstration, let's assume a mock CPU usage of 0.5 (50%)
	cpuUsage := 0.5

	// Calculate response time
	responseTime := time.Since(start)

	// Estimate energy consumption
	energy := EnergyEstimate(responseTime, cpuUsage)
	log.Printf("The user is %s", user)

	metrics := models.Metrics{
		APIEndpoint:       endpoint,
		UserId:            user,
		RequestSize:       len(endpoint), // Request size can include headers, etc.
		ResponseSize:      len(body),
		ResponseTime:      responseTime,
		Timestamp:         time.Now(),
		EnergyConsumption: energy,
	}

	log.Printf("The user is %s", metrics.UserId)
	// Debugging print statements
	fmt.Printf("Collected metrics: %+v\n", metrics)

	return metrics
}
