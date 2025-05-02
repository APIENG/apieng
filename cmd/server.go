package main

import (
	"log"
	"net/http"

	"github.com/APIENG/apieng/internal/handlers"
)

func main() {
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	mux.HandleFunc("/", handlers.LandingHandler)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request for /metrics with method: %s", r.Method)
		handlers.MetricsHandler(w, r)
	})

	// Add dashboard route
	mux.HandleFunc("/dashboard", handlers.DashboardHandler)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
