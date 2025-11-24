package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PredictionClient handles communication with the AI prediction service
type PredictionClient struct {
	serviceURL string
	client     *http.Client
}

// NewPredictionClient creates a new AI prediction client
func NewPredictionClient(serviceURL string) *PredictionClient {
	return &PredictionClient{
		serviceURL: serviceURL,
		client: &http.Client{
			Timeout: 500 * time.Millisecond, // Fast timeout for real-time decisions
		},
	}
}

// PredictionRequest represents a request to the AI service
type PredictionRequest struct {
	Features []float64 `json:"features"`
	RouteID  string    `json:"route_id"`
}

// PredictionResponse represents a response from the AI service
type PredictionResponse struct {
	RouteID            string  `json:"route_id"`
	PredictedLatencyMs float64 `json:"predicted_latency_ms"`
	PredictedJitterMs  float64 `json:"predicted_jitter_ms"`
	ConfidenceScore    float64 `json:"confidence_score"`
}

// RoutesRequest represents a request to compare multiple routes
type RoutesRequest struct {
	Routes map[string][]float64 `json:"routes"`
}

// RoutesResponse represents a response for route comparison
type RoutesResponse struct {
	BestRoute   string                        `json:"best_route"`
	Predictions map[string]PredictionResponse `json:"predictions"`
	Ranking     []string                      `json:"ranking"`
}

// GetPrediction gets a prediction for a single route
func (c *PredictionClient) GetPrediction(routeID string, features []float64) (*PredictionResponse, error) {
	reqBody := PredictionRequest{
		RouteID:  routeID,
		Features: features,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.serviceURL+"/predict", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("service returned status: %d", resp.StatusCode)
	}

	var result PredictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetBestRoute compares multiple routes and returns the best one
func (c *PredictionClient) GetBestRoute(routes map[string][]float64) (string, error) {
	reqBody := RoutesRequest{
		Routes: routes,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.serviceURL+"/predict/routes", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("service returned status: %d", resp.StatusCode)
	}

	var result RoutesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.BestRoute, nil
}
