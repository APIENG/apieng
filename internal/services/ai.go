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
	aiMetrics := ""
	aiErr := callGeminiAPI(requestSize, responseSize, responseTime.Seconds(), &aiMetrics)
	if aiErr != nil {
		fmt.Printf("Warning: Gemini API failed: %v - using fallback explanation\n", aiErr)
		aiMetrics = fmt.Sprintf("Response in %.2f seconds with %d bytes transferred. Unable to generate AI insights at this time.", responseTime.Seconds(), responseSize)
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

func callGeminiAPI(requestSize, responseSize int, responseTime float64, result *string) error {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found, relying on system env")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY is not set")
	}

	// Calculate derived metrics (with safe defaults)
	totalDataKB := float64(requestSize+responseSize) / 1024.0
	dataRateKBps := 0.0
	if responseTime > 0 {
		dataRateKBps = totalDataKB / responseTime
	}
	requestResponseRatio := 1.0
	if requestSize > 0 {
		requestResponseRatio = float64(responseSize) / float64(requestSize)
	}
	responseSizeMB := float64(responseSize) / (1024 * 1024)

	prompt := fmt.Sprintf(`Analyze this API endpoint's performance:
- Request: %d bytes
- Response: %d bytes
- Duration: %.2f seconds

Provide a brief, actionable performance analysis covering:
1. Overall data efficiency (total %.2f KB transferred in %.2f seconds = %.2f KB/s throughput)
2. Response size impact (%.2f MB response is %s size relative to request)
3. Performance assessment: Is this fast, normal, or slow for the response size?
4. One specific optimization recommendation

Keep it concise and practical. Focus on what this means for real-world usage.`,
		requestSize, responseSize, responseTime,
		totalDataKB, responseTime, dataRateKBps,
		responseSizeMB, getResponseSizeAssessment(requestResponseRatio),
	)

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
		return fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", apiKey), bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("request creation failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("response read failed: %v", err)
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
		return fmt.Errorf("response unmarshal failed: %v", err)
	}

	if len(responseBody.Candidates) == 0 || len(responseBody.Candidates[0].Content.Parts) == 0 {
		return fmt.Errorf("no valid candidates in Gemini response")
	}

	*result = responseBody.Candidates[0].Content.Parts[0].Text
	return nil
}

// getResponseSizeAssessment returns a human-readable assessment of response size relative to request
func getResponseSizeAssessment(ratio float64) string {
	switch {
	case ratio < 0.5:
		return "smaller"
	case ratio < 1.5:
		return "comparable"
	case ratio < 5:
		return "moderately larger"
	default:
		return "significantly larger"
	}
}
