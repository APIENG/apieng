//go:build ignore

// Standalone scratch script for testing the Gemini API. Run directly with
// `go run api_testing.go`. Excluded from the normal build.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// API endpoint
const apiURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent"

// RequestBody struct to send data to Gemini API
type RequestBody struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

// ResponseBody struct to parse Gemini API response
type ResponseBody struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// generateMetricsRequest constructs the prompt and sends it to the Gemini API
func generateMetricsRequest(requestSize int, responseSize int, responseTime float64) (string, error) {
	// Get API key from environment variable
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	// Construct the AI prompt
	prompt := fmt.Sprintf(
		`I have API data for an endpoint:
Request Size: %d bytes
Response Size: %d bytes
Response Time: %.2f seconds

Please calculate the following metrics:
1.  Total data transferred (request + response) in kilobytes.
2.  Data transfer rate (total data transferred / response time) in kilobytes per second.
3.  Request size to response size ratio.
4.  Any other metric you think might be related to the endpoint's computational load. Explain why you think this new metric is relevant.

Format your response as a JSON object with keys: 'total_data_kb', 'data_rate_kbps', 'request_response_ratio', 'additional_metric_name', 'additional_metric_value', 'explanation'.`,
		requestSize, responseSize, responseTime,
	)

	// Create request payload
	requestBody := RequestBody{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: prompt},
				},
			},
		},
	}

	// Convert request body to JSON
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error encoding request JSON: %v", err)
	}

	// Make HTTP request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s?key=%s", apiURL, apiKey), bytes.NewBuffer(requestJSON))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	// Parse response JSON
	var responseBody ResponseBody
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return "", fmt.Errorf("error decoding response JSON: %v", err)
	}

	// Extract and return AI response
	if len(responseBody.Candidates) > 0 && len(responseBody.Candidates[0].Content.Parts) > 0 {
		return responseBody.Candidates[0].Content.Parts[0].Text, nil
	}
	return "", fmt.Errorf("no response from API")
}

func main() {
	// Test the function with sample values
	requestSize := 2048  // in bytes
	responseSize := 8192 // in bytes
	responseTime := 0.75 // in seconds

	// Call function to get AI-generated metrics
	response, err := generateMetricsRequest(requestSize, responseSize, responseTime)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Print the AI response
	fmt.Println("AI Response:", response)
}
