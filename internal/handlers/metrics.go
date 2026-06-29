package handlers

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"

	"github.com/APIENG/apieng/internal/db"
	"github.com/APIENG/apieng/internal/models"
	"github.com/APIENG/apieng/internal/services"
)

// MetricsTemplateData represents the data to be passed to the HTML template.
type MetricsTemplateData struct {
	Metrics      []models.Metrics
	CurrentPage  int
	TotalPages   int
	HasPrevious  bool
	HasNext      bool
	PreviousPage int
	NextPage     int
}

type RequestBody struct {
	APIEndpoint string `json:"apiEndpoint"`
}

func FetchMetrics(db *sql.DB, UserId string) ([]models.Metrics, error) {
	rows, err := db.Query(`SELECT id, api_endpoint, request_size, response_size, response_time, timestamp, energy_consumption, user_id, status, method, explanation FROM metrics WHERE user_id = ?`, UserId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metricsList []models.Metrics
	for rows.Next() {
		var m models.Metrics
		var responseTime int64
		err := rows.Scan(&m.ID, &m.APIEndpoint, &m.RequestSize, &m.ResponseSize, &responseTime, &m.Timestamp, &m.EnergyConsumption, &m.UserId, &m.Status, &m.Method, &m.Explanation)
		if err != nil {
			return nil, err
		}
		m.ResponseTime = time.Duration(responseTime) * time.Millisecond
		//fmt.Printf("Metric: %+v\n", m)
		metricsList = append(metricsList, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return metricsList, nil
}

func FetchMetricsByID(db *sql.DB, userId, metricId string) (*models.Metrics, error) {
	idInt, erra := strconv.Atoi(metricId)
	if erra != nil {
		return nil, fmt.Errorf("invalid metric ID: %v", erra)
	}
	row := db.QueryRow(`
		SELECT id, api_endpoint, request_size, response_size, response_time, timestamp,
		       energy_consumption, user_id, status, method, explanation
		FROM metrics
		WHERE user_id = ? AND id = ?
	`, userId, idInt)

	var m models.Metrics
	var responseTime int64

	err := row.Scan(
		&m.ID,
		&m.APIEndpoint,
		&m.RequestSize,
		&m.ResponseSize,
		&responseTime,
		&m.Timestamp,
		&m.EnergyConsumption,
		&m.UserId,
		&m.Status,
		&m.Method,
		&m.Explanation,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}

	m.ResponseTime = time.Duration(responseTime) * time.Millisecond
	return &m, nil
}

// Handler to display the metrics page
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	dbConn, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}

	token := context.Get(r, "user")
	strToken, _ := token.(string)
	allMetrics, err := FetchMetrics(dbConn, strToken)
	if err != nil {
		http.Error(w, "Unable to fetch metrics", http.StatusInternalServerError)
		return
	}

	// Pagination: 10 items per page
	pageSize := 10
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	totalPages := (len(allMetrics) + pageSize - 1) / pageSize
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(allMetrics) {
		end = len(allMetrics)
	}

	var pageMetrics []models.Metrics
	if start < len(allMetrics) {
		pageMetrics = allMetrics[start:end]
	}

	tmpl, err := template.ParseFiles("templates/metrics.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	templateData := MetricsTemplateData{
		Metrics:      pageMetrics,
		CurrentPage:  page,
		TotalPages:   totalPages,
		HasPrevious:  page > 1,
		HasNext:      page < totalPages,
		PreviousPage: page - 1,
		NextPage:     page + 1,
	}

	err = tmpl.Execute(w, templateData)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

//Handler to Handle the metrics page with a specific ID
func EachMetricsHandler(w http.ResponseWriter, r *http.Request) {
	dbConn, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	token := context.Get(r, "user")
	strToken, _ := token.(string)

	// Pass the ID to your fetch logic if needed
	metrics, err := FetchMetricsByID(dbConn, strToken, id)
	if err != nil {
		http.Error(w, "Unable to fetch metrics", http.StatusInternalServerError)
		return
	}

	//fmt.Printf("Fetched metrics for ID %s: %+v\n", id, metrics)

	tmpl, err := template.ParseFiles("templates/each_metric.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, metrics)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// Handler to process csv endpoint
func ExportMetricsCSVHandler(w http.ResponseWriter, r *http.Request) {
	dbConn, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}

	token := context.Get(r, "user")
	strToken, _ := token.(string)

	metrics, err := FetchMetrics(dbConn, strToken)
	if err != nil {
		http.Error(w, "Unable to fetch metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment;filename=metrics.csv")
	w.Header().Set("Content-Type", "text/csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	writer.Write([]string{
		"ID", "APIEndpoint", "UserId", "Method", "Status",
		"RequestSize", "ResponseSize", "ResponseTime (ms)", "Timestamp", "EnergyConsumption", "Explanation",
	})

	// Write rows
	for _, m := range metrics {
		writer.Write([]string{
			strconv.Itoa(m.ID),
			m.APIEndpoint,
			m.UserId,
			m.Method,
			strconv.Itoa(m.Status),
			strconv.Itoa(m.RequestSize),
			strconv.Itoa(m.ResponseSize),
			strconv.FormatInt(m.ResponseTime.Milliseconds(), 10),
			m.Timestamp.Format(time.RFC3339),
			fmt.Sprintf("%.4f", m.EnergyConsumption),
			m.Explanation,
		})
	}
}

// Handler to process csv endpoint for a particular Endpoint
func EachExportMetricsCSVHandler(w http.ResponseWriter, r *http.Request) {
	dbConn, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	token := context.Get(r, "user")
	strToken, _ := token.(string)

	metric, err := FetchMetricsByID(dbConn, strToken, id)
	if err != nil {
		http.Error(w, "Unable to fetch metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment;filename=metrics.csv")
	w.Header().Set("Content-Type", "text/csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{
		"ID", "APIEndpoint", "UserId", "Method", "Status",
		"RequestSize", "ResponseSize", "ResponseTime (ms)", "Timestamp", "EnergyConsumption", "Explanation",
	})

	writer.Write([]string{
		strconv.Itoa(metric.ID),
		metric.APIEndpoint,
		metric.UserId,
		metric.Method,
		strconv.Itoa(metric.Status),
		strconv.Itoa(metric.RequestSize),
		strconv.Itoa(metric.ResponseSize),
		strconv.FormatInt(metric.ResponseTime.Milliseconds(), 10),
		metric.Timestamp.Format(time.RFC3339),
		fmt.Sprintf("%.4f", metric.EnergyConsumption),
		metric.Explanation,
	})

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

	// Validate URL format
	if !isValidURL(apiEndpoint) {
		http.Error(w, "Invalid API endpoint URL format", http.StatusBadRequest)
		return
	}

	// Measure the API and store the metrics
	dr, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}

	cookie, err := r.Cookie("user_id")
	if err != nil || cookie == nil {
		http.Error(w, "User session not found", http.StatusUnauthorized)
		return
	}

	metrics := services.MeasureAPIWithAI(apiEndpoint, cookie.Value)
	err = db.StoreMetrics(dr, metrics)
	if err != nil {
		http.Error(w, "Error storing metrics", http.StatusInternalServerError)
		return
	}

	// Return JSON for AJAX requests, redirect for regular form submissions
	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		// Escape HTML in explanation to prevent XSS
		metrics.Explanation = escapeHTML(metrics.Explanation)
		response := map[string]interface{}{
			"APIEndpoint":       metrics.APIEndpoint,
			"UserId":            metrics.UserId,
			"Method":            metrics.Method,
			"Status":            metrics.Status,
			"RequestSize":       metrics.RequestSize,
			"ResponseSize":      metrics.ResponseSize,
			"ResponseTime":      metrics.ResponseTime.Milliseconds(),
			"Timestamp":         metrics.Timestamp,
			"EnergyConsumption": metrics.EnergyConsumption,
			"Explanation":       metrics.Explanation,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	http.Redirect(w, r, "/metrics", http.StatusSeeOther)
}

func ApiMeasureHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody RequestBody

	// Decode the JSON body into struct
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	apiEndpoint := reqBody.APIEndpoint

	// apiEndpoint := r.FormValue("apiEndpoint")
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

	token := context.Get(r, "user")
	strToken, _ := token.(string)
	metrics := services.MeasureAPIWithAI(apiEndpoint, strToken)
	//metrics := services.MeasureAPI(apiEndpoint, strToken)
	err = db.StoreMetrics(dr, metrics)
	if err != nil {
		http.Error(w, "Error storing metrics", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(metrics)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}

	// Redirect back to the metrics page
	//http.Redirect(w, r, "/api/metrics", http.StatusSeeOther)
}

// API handler to return metrics in JSON format
func APIMetricsHandler(w http.ResponseWriter, r *http.Request) {
	dbConn, err := db.InitializeDB()
	if err != nil {
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}

	token := context.Get(r, "user")
	strToken, _ := token.(string)
	metrics, err := FetchMetrics(dbConn, strToken)
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

// isValidURL validates that the string is a properly formatted URL
func isValidURL(endpoint string) bool {
	if endpoint == "" {
		return false
	}
	// Must start with http:// or https://
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		return false
	}
	// Parse as URL to validate structure
	_, err := url.Parse(endpoint)
	return err == nil
}

// escapeHTML safely escapes HTML to prevent XSS
func escapeHTML(s string) string {
	return html.EscapeString(s)
}
