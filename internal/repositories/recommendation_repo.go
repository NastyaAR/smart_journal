package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"blockchain_project/internal/models"

	"github.com/jackc/pgx/v5"
)

type RecommendationRepository struct {
	*PostgresRepository
}

func NewRecommendationRepository(repo *PostgresRepository) *RecommendationRepository {
	return &RecommendationRepository{PostgresRepository: repo}
}

func (r *RecommendationRepository) CreateRecommendation(ctx context.Context, studentID int, recommendation *models.AIRecommendationResponse) (*models.StoredRecommendation, error) {
	payload, err := json.Marshal(recommendation)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recommendation: %w", err)
	}

	stored := &models.StoredRecommendation{
		StudentID: studentID,
		Payload:   recommendation,
	}
	query := `INSERT INTO student_recommendations (student_id, payload) VALUES ($1, $2::jsonb) RETURNING id, created_at`
	if err := r.pool.QueryRow(ctx, query, studentID, string(payload)).Scan(&stored.ID, &stored.CreatedAt); err != nil {
		return nil, fmt.Errorf("failed to create recommendation: %w", err)
	}

	return stored, nil
}

func (r *RecommendationRepository) GetLatestByStudentID(ctx context.Context, studentID int) (*models.StoredRecommendation, error) {
	var stored models.StoredRecommendation
	var payload []byte
	query := `
		SELECT id, student_id, payload, created_at
		FROM student_recommendations
		WHERE student_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT 1`
	err := r.pool.QueryRow(ctx, query, studentID).Scan(&stored.ID, &stored.StudentID, &payload, &stored.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("recommendation not found")
		}
		return nil, fmt.Errorf("failed to get recommendation: %w", err)
	}

	var recommendation models.AIRecommendationResponse
	if err := json.Unmarshal(payload, &recommendation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recommendation: %w", err)
	}
	stored.Payload = &recommendation

	return &stored, nil
}
