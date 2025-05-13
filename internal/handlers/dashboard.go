package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"time"

	"github.com/APIENG/apieng/internal/db"
	"github.com/APIENG/apieng/internal/models"
	"github.com/gorilla/context"
)

type DashboardData struct {
	Metrics []models.Metrics
	Stats   struct {
		TotalRequests   int
		AvgResponseTime float64
		EnergyUsage     float64
		ActiveEndpoints int
	}
}

func FetchMetricsLast24Hours(db *sql.DB, userID string) ([]models.Metrics, error) {
	query := `
		SELECT api_endpoint, user_id, method, status, request_size, response_size,
		       response_time, timestamp, energy_consumption
		FROM metrics
		WHERE user_id = ? AND timestamp >= datetime('now', '-24 hours')
		ORDER BY timestamp DESC;
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metricsList []models.Metrics
	for rows.Next() {
		var m models.Metrics
		var responseTimeSec float64

		err := rows.Scan(
			&m.APIEndpoint, &m.UserId, &m.Method, &m.Status,
			&m.RequestSize, &m.ResponseSize, &responseTimeSec,
			&m.Timestamp, &m.EnergyConsumption,
		)
		if err != nil {
			return nil, err
		}

		m.ResponseTime = time.Duration(responseTimeSec * float64(time.Second))
		metricsList = append(metricsList, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return metricsList, nil
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Connect to DB
	db, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Extract user ID from context
	token := context.Get(r, "user")
	strToken, _ := token.(string)
	metrics, err := FetchMetricsLast24Hours(db, strToken)
	if err != nil {
		http.Error(w, "Unable to fetch metrics", http.StatusInternalServerError)
		return
	}

	// Calculate dashboard stats
	var totalRequests, totalEndpoints int
	var totalResponseTime float64
	var totalEnergy float64
	endpointSet := make(map[string]struct{})

	for _, m := range metrics {
		totalRequests++
		totalResponseTime += m.ResponseTime.Seconds()
		totalEnergy += m.EnergyConsumption
		endpointSet[m.APIEndpoint] = struct{}{}
	}
	totalEndpoints = len(endpointSet)

	var avgResponse float64
	if totalRequests > 0 {
		avgResponse = totalResponseTime / float64(totalRequests)
	}

	// Prepare data
	data := DashboardData{
		Metrics: metrics,
	}
	data.Stats.TotalRequests = totalRequests
	data.Stats.AvgResponseTime = avgResponse
	data.Stats.EnergyUsage = totalEnergy
	data.Stats.ActiveEndpoints = totalEndpoints

	// Render template
	tmpl := template.Must(template.ParseFiles("templates/dashboard.html"))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
