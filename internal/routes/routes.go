package routes

import (
	"github.com/APIENG/apieng/internal/handlers"
	"github.com/APIENG/apieng/pkg"
	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	router := mux.NewRouter()

	// Users Handler
	router.HandleFunc("/", handlers.LoginHandler).Methods("GET")
	router.HandleFunc("/signup", handlers.SignUpHandler).Methods("GET")
	router.HandleFunc("/login", handlers.LoginusersHandler).Methods("POST")
	router.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	router.HandleFunc("/users", pkg.AuthorizeCookie(handlers.UsersHandler)).Methods("GET")
	router.HandleFunc("/logout", pkg.AuthorizeCookie(handlers.LogoutHandler)).Methods("GET")
	router.HandleFunc("/api/users", handlers.APIusersHandler).Methods("GET")

	//Metrics handler
	router.HandleFunc("/metrics", pkg.AuthorizeCookie(handlers.MetricsHandler)).Methods("GET")
	router.HandleFunc("/measure", pkg.AuthorizeCookie(handlers.MeasureHandler)).Methods("POST")
	router.HandleFunc("/api/metrics", pkg.Authorize(handlers.APIMetricsHandler)).Methods("GET") // New API endpoint

	return router
}
