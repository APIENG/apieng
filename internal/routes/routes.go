package routes

import (
	"net/http"

	"github.com/APIENG/apieng/internal/handlers"
	"github.com/APIENG/apieng/pkg"
	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	router := mux.NewRouter()

	fs := http.FileServer(http.Dir("static"))
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	router.HandleFunc("/", handlers.LandingHandler).Methods("GET")
	router.HandleFunc("/login", handlers.LoginHandler).Methods("GET")
	router.HandleFunc("/signup", handlers.SignUpHandler).Methods("GET")
	router.HandleFunc("/login", handlers.LoginusersHandler).Methods("POST")
	router.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	router.HandleFunc("/users", pkg.AuthorizeCookie(handlers.UsersHandler)).Methods("GET")
	router.HandleFunc("/generate", pkg.AuthorizeCookie(handlers.GenerateKey)).Methods("POST")
	router.HandleFunc("/logout", pkg.AuthorizeCookie(handlers.LogoutHandler)).Methods("GET")
	router.HandleFunc("/api/users", handlers.APIusersHandler).Methods("GET")

	//Metrics handler
	router.HandleFunc("/metrics", pkg.AuthorizeCookie(handlers.MetricsHandler)).Methods("GET")
	router.HandleFunc("/metrics/{id}", pkg.AuthorizeCookie(handlers.EachMetricsHandler)).Methods("GET")
	router.HandleFunc("/metrics/export/download", pkg.AuthorizeCookie(handlers.ExportMetricsCSVHandler)).Methods("GET")
	router.HandleFunc("/dashboard", pkg.AuthorizeCookie(handlers.DashboardHandler)).Methods("GET")
	router.HandleFunc("/api/measure", handlers.AuthorizeAPI(handlers.ApiMeasureHandler)).Methods("POST")
	router.HandleFunc("/measure", pkg.AuthorizeCookie(handlers.MeasureHandler)).Methods("POST")
	router.HandleFunc("/api/metrics", handlers.AuthorizeAPI(handlers.APIMetricsHandler)).Methods("GET") // New API endpoint

	return router
}
