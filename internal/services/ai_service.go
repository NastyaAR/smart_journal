package services

import (
	"context"
	"fmt"
)

type AIService struct{}

func NewAIService() *AIService {
	return &AIService{}
}

func (a *AIService) GetRecommendations(ctx context.Context, studentID int) (string, error) {
	return fmt.Sprintf("Студент %d: Рекомендуется уделить больше внимания математике и принять участие в научном кружке.", studentID), nil
}
