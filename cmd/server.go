package main

import (
	"log"
	"net/http"

	"github.com/APIENG/apieng/internal/routes"
)

func main() {
	router := routes.SetupRouter()

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
