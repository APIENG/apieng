package handlers

import (
	"html/template"
	"net/http"
)

type Metric struct {
	Name  string
	Value float64
}

type DashboardData struct {
	Metrics []Metric
	Stats   struct {
		TotalRequests   int
		AvgResponseTime float64
		EnergyUsage     float64
		ActiveEndpoints int
	}
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Populate your dashboard data here
	data := DashboardData{
		// Add your metrics data
	}

	tmpl := template.Must(template.ParseFiles("templates/dashboard.html"))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
