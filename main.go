//go:build ignore

// Legacy pre-refactor entry point (PostgreSQL-based). The real entry point is
// cmd/server.go. Excluded from the build to avoid a duplicate main declaration.
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"text/template"
	"time"
	// _ "github.com/lib/pq"
)

type Metrics struct {
	APIEndpoint       string
	RequestSize       int64
	ResponseSize      int64
	ResponseTime      time.Duration
	Timestamp         time.Time
	EnergyConsumption float64
}

type MetricsTemplateData struct {
	Metrics []Metrics
}

func InitializeDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgres://username:password@localhost:5432/dbname?sslmode=disable")
	if err != nil {
		return nil, err
	}
	return db, nil
}

// func FetchMetrics(db *sql.DB) ([]Metrics, error) {
func FetchMetrics(db *sql.DB) ([]Metrics, error) {
	rows, err := db.Query(`SELECT api_endpoint, request_size, response_size, response_time, timestamp, energy_consumption FROM metrics`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metricsList []Metrics
	for rows.Next() {
		var m Metrics
		var responseTime int64
		err := rows.Scan(&m.APIEndpoint, &m.RequestSize, &m.ResponseSize, &responseTime, &m.Timestamp, &m.EnergyConsumption)
		if err != nil {
			return nil, err
		}
		m.ResponseTime = time.Duration(responseTime) * time.Millisecond
		metricsList = append(metricsList, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return metricsList, nil
}

// Handler to display the metrics page
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	metrics, err := FetchMetrics(db)
	if err != nil {
		http.Error(w, "Unable to fetch metrics", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/metrics.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, MetricsTemplateData{Metrics: metrics})
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// Handler to process API endpoint form submission and measure API metrics
func measureHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiEndpoint := r.FormValue("apiEndpoint")
	if apiEndpoint == "" {
		http.Error(w, "API endpoint is required", http.StatusBadRequest)
		return
	}

	// Measure the API and store the metrics
	db, err := InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	metrics := MeasureAPI(apiEndpoint)
	err = StoreMetrics(db, metrics)
	if err != nil {
		http.Error(w, "Error storing metrics", http.StatusInternalServerError)
		return
	}

	// Redirect back to the metrics page
	http.Redirect(w, r, "/metrics", http.StatusSeeOther)
}

// MeasureAPI performs API measurements and returns metrics
func MeasureAPI(apiEndpoint string) Metrics {
	start := time.Now()
	resp, err := http.Get(apiEndpoint)
	if err != nil {
		return Metrics{APIEndpoint: apiEndpoint}
	}
	defer resp.Body.Close()

	responseTime := time.Since(start)
	// Simplified energy consumption calculation
	energyConsumption := float64(responseTime.Milliseconds()) * 0.001

	return Metrics{
		APIEndpoint:       apiEndpoint,
		RequestSize:       resp.Request.ContentLength,
		ResponseSize:      resp.ContentLength,
		ResponseTime:      responseTime,
		Timestamp:         time.Now(),
		EnergyConsumption: energyConsumption,
	}
}

// API handler to return metrics in JSON format
func apiMetricsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	metrics, err := FetchMetrics(db)
	if err != nil {
		http.Error(w, "Unable to fetch metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(metrics)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
}

// StoreMetrics stores the metrics data in the database
func StoreMetrics(db *sql.DB, m Metrics) error {
	_, err := db.Exec(`
		INSERT INTO metrics (api_endpoint, request_size, response_size, response_time, timestamp, energy_consumption)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		m.APIEndpoint, m.RequestSize, m.ResponseSize, m.ResponseTime.Milliseconds(), m.Timestamp, m.EnergyConsumption)
	return err
}

func main() {
	http.HandleFunc("/metrics", metricsHandler)
	http.HandleFunc("/measure", measureHandler)
	http.HandleFunc("/api/metrics", apiMetricsHandler) // New API endpoint

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
