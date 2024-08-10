package routes

import (
	"github.com/Taiwrash/apieng/internal/handlers"
	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	router := mux.NewRouter()

	// Use the handlers directly
	router.HandleFunc("/metrics", handlers.MetricsHandler).Methods("GET")
	router.HandleFunc("/measure", handlers.MeasureHandler).Methods("POST")
	router.HandleFunc("/api/metrics", handlers.APIMetricsHandler).Methods("GET") // New API endpoint

	return router
}
