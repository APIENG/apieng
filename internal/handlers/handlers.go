package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"text/template"
	"time"

	"github.com/Taiwrash/apieng/internal/db"
	"github.com/Taiwrash/apieng/internal/models"
	"github.com/Taiwrash/apieng/internal/services"
)

// MetricsTemplateData represents the data to be passed to the HTML template.
type MetricsTemplateData struct {
	Metrics []models.Metrics
}

func FetchMetrics(db *sql.DB) ([]models.Metrics, error) {
	rows, err := db.Query(`SELECT api_endpoint, request_size, response_size, response_time, timestamp, energy_consumption FROM metrics`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metricsList []models.Metrics
	for rows.Next() {
		var m models.Metrics
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
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := db.InitializeDB()
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
func MeasureHandler(w http.ResponseWriter, r *http.Request) {
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
	dr, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer dr.Close()

	metrics := services.MeasureAPI(apiEndpoint)
	err = db.StoreMetrics(dr, metrics)
	if err != nil {
		http.Error(w, "Error storing metrics", http.StatusInternalServerError)
		return
	}

	// Redirect back to the metrics page
	http.Redirect(w, r, "/api/metrics", http.StatusSeeOther)
}

// API handler to return metrics in JSON format
func APIMetricsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := db.InitializeDB()
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
