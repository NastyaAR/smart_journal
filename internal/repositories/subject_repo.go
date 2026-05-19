package repositories

import (
	"context"
	"fmt"

	"blockchain_project/internal/models"

	"github.com/jackc/pgx/v5"
)

type SubjectRepository struct {
	*PostgresRepository
}

func NewSubjectRepository(repo *PostgresRepository) *SubjectRepository {
	return &SubjectRepository{PostgresRepository: repo}
}

func (r *SubjectRepository) CreateSubject(ctx context.Context, subject *models.Subject) error {
	query := `INSERT INTO subjects (name) VALUES ($1) RETURNING id`
	return r.pool.QueryRow(ctx, query, subject.Name).Scan(&subject.ID)
}

func (r *SubjectRepository) GetSubjectByID(ctx context.Context, id int) (*models.Subject, error) {
	var subject models.Subject
	query := `SELECT id, name FROM subjects WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&subject.ID, &subject.Name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("subject not found")
		}
		return nil, fmt.Errorf("failed to get subject: %w", err)
	}
	return &subject, nil
}