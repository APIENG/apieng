package pkg

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string    `json:"error"`
	Message string    `json:"message"`
	Code    int       `json:"code"`
	Time    time.Time `json:"timestamp"`
}

// SendError sends a standardized error response
func SendError(w http.ResponseWriter, statusCode int, errorMsg string, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   errorMsg,
		Message: details,
		Code:    statusCode,
		Time:    time.Now(),
	}

	json.NewEncoder(w).Encode(response)
	log.Printf("Error %d: %s - %s", statusCode, errorMsg, details)
}

// SendJSONError sends an error response for AJAX requests
func SendJSONError(w http.ResponseWriter, statusCode int, errorMsg string) {
	SendError(w, statusCode, errorMsg, "")
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Data   interface{} `json:"data"`
	Status string      `json:"status"`
	Time   time.Time   `json:"timestamp"`
}

// SendSuccess sends a standardized success response
func SendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := SuccessResponse{
		Data:   data,
		Status: "success",
		Time:   time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}
