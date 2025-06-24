package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/APIENG/apieng/internal/models"
	"github.com/joho/godotenv"
)

// MeasureAPI collects metrics for a given API endpoint and estimates energy consumption.
func MeasureAPIWithAI(endpoint string, user string) models.Metrics {
	start := time.Now()
	resp, err := http.Get(endpoint)
	if err != nil {
		fmt.Println("Error making request:", err)
		return models.Metrics{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return models.Metrics{}
	}

	responseTime := time.Since(start)
	requestSize := len(endpoint)
	responseSize := len(body)

	// Simulated CPU usage
	cpuUsage := 0.5
	energy := EnergyEstimate(responseTime, cpuUsage)

	// Call Gemini to generate advanced metrics
	aiMetrics, err := callGeminiAPI(requestSize, responseSize, responseTime.Seconds())
	if err != nil {
		fmt.Println("Gemini error:", err)
	}

	metrics := models.Metrics{
		APIEndpoint:       endpoint,
		UserId:            user,
		RequestSize:       requestSize,
		ResponseSize:      responseSize,
		ResponseTime:      responseTime,
		Timestamp:         time.Now(),
		Method:            "GET",
		Status:            resp.StatusCode,
		EnergyConsumption: energy,
		Explanation:       aiMetrics,
	}

	fmt.Printf("Collected metrics: %+v\n", metrics)
	return metrics
}

func callGeminiAPI(requestSize, responseSize int, responseTime float64) (string, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found, relying on system env")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY is not set")
	}

	prompt := fmt.Sprintf(`
I have API data for an endpoint:
Request Size: %d bytes
Response Size: %d bytes
Response Time: %.2f seconds

Please provide a clear explanation of the following:

1. The total data transferred (sum of request and response sizes) in kilobytes.
2. The data transfer rate in kilobytes per second.
3. The ratio of request size to response size.
4. Include any other metric you consider relevant to understanding the endpoint's computational behavior and explain its significance.

Respond only with a plain text explanation. Do not format as JSON or list keys. Just give a concise, well-written paragraph that includes all the information.
`, requestSize, responseSize, responseTime)

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", apiKey), bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("request creation failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("response read failed: %v", err)
	}
	//fmt.Println("Gemini API Response:\n", string(body))

	// Parse response
	var responseBody struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return "", fmt.Errorf("response unmarshal failed: %v", err)
	}

	if len(responseBody.Candidates) == 0 || len(responseBody.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no valid candidates in Gemini response")
	}

	rawText := responseBody.Candidates[0].Content.Parts[0].Text

	// Extract and parse AI JSON

	return rawText, nil
}
