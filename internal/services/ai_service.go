package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"blockchain_project/internal/models"
)

type AIService struct {
	baseURL    string
	httpClient *http.Client
}

func NewAIService(baseURL string) *AIService {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8000"
	}

	return &AIService{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (a *AIService) GetRecommendations(ctx context.Context, req *models.AIRecommendationRequest) (*models.AIRecommendationResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AI request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/get_recommendations", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create AI request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call AI service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("AI service returned status %d", resp.StatusCode)
	}

	var recommendation models.AIRecommendationResponse
	if err := json.NewDecoder(resp.Body).Decode(&recommendation); err != nil {
		return nil, fmt.Errorf("failed to decode AI response: %w", err)
	}

	return &recommendation, nil
}
