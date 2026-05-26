package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
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
		baseURL = "http://llm:8000" //"http://127.0.0.1:8000"
	}

	return &AIService{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5000 * time.Second,
		},
	}
}

func (a *AIService) GetRecommendations(ctx context.Context, req *models.AIRecommendationRequest) (*models.AIRecommendationResponse, error) {

	log.Printf("[AI] Input: ID=%s, Name=%q, Surname=%q, GradesCount=%d",
		req.StudentID, req.StudentName, req.StudentSurname, len(req.Grades))

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		log.Printf("[AI] Marshal error: %v", err)
		return nil, err
	}

	log.Printf("[AI] Request JSON: %s", string(bodyBytes))
	targetURL := a.baseURL + "/get_recommendations"
	log.Printf("[AI] Calling URL: %s", targetURL)
	log.Printf("[AI] Timeout: %v", a.httpClient.Timeout)

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

type ChatRequest struct {
	Message string `json:"message"`
	Context string `json:"context,omitempty"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

func (a *AIService) Chat(ctx context.Context, message string, context string) (string, error) {
	reqBody := ChatRequest{
		Message: message,
		Context: context,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal chat request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/chat", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create chat request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call chat service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("chat service returned status %d", resp.StatusCode)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode chat response: %w", err)
	}

	return chatResp.Response, nil
}
