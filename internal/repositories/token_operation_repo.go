package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"blockchain_project/internal/models"
)

type TokenOperationRepository struct {
	*PostgresRepository
}

func NewTokenOperationRepository(repo *PostgresRepository) *TokenOperationRepository {
	return &TokenOperationRepository{PostgresRepository: repo}
}

func (r *TokenOperationRepository) Create(ctx context.Context, operation *models.TokenOperation) error {
	query := `
		INSERT INTO token_operations (student_id, teacher_id, amount, operation_type, reason)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	return r.pool.QueryRow(
		ctx,
		query,
		operation.StudentID,
		operation.TeacherID,
		operation.Amount,
		operation.OperationType,
		operation.Reason,
	).Scan(&operation.ID, &operation.CreatedAt)
}

func (r *TokenOperationRepository) GetByStudentID(ctx context.Context, studentID int) ([]*models.TokenOperation, error) {
	query := `
		SELECT o.id, o.student_id, s.name, o.teacher_id, COALESCE(t.name, ''),
		       o.amount, o.operation_type, COALESCE(o.reason, ''), o.created_at
		FROM token_operations o
		JOIN students s ON s.id = o.student_id
		LEFT JOIN teachers t ON t.id = o.teacher_id
		WHERE o.student_id = $1
		ORDER BY o.created_at DESC, o.id DESC`
	return r.queryOperations(ctx, query, studentID)
}

func (r *TokenOperationRepository) GetByGroupID(ctx context.Context, groupID int) ([]*models.TokenOperation, error) {
	query := `
		SELECT o.id, o.student_id, s.name, o.teacher_id, COALESCE(t.name, ''),
		       o.amount, o.operation_type, COALESCE(o.reason, ''), o.created_at
		FROM token_operations o
		JOIN students s ON s.id = o.student_id
		LEFT JOIN teachers t ON t.id = o.teacher_id
		WHERE s.group_id = $1
		ORDER BY o.created_at DESC, o.id DESC`
	return r.queryOperations(ctx, query, groupID)
}

func (r *TokenOperationRepository) queryOperations(ctx context.Context, query string, args ...any) ([]*models.TokenOperation, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query token operations: %w", err)
	}
	defer rows.Close()

	var operations []*models.TokenOperation
	for rows.Next() {
		var operation models.TokenOperation
		var teacherID sql.NullInt64
		if err := rows.Scan(
			&operation.ID,
			&operation.StudentID,
			&operation.StudentName,
			&teacherID,
			&operation.TeacherName,
			&operation.Amount,
			&operation.OperationType,
			&operation.Reason,
			&operation.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan token operation: %w", err)
		}
		if teacherID.Valid {
			value := int(teacherID.Int64)
			operation.TeacherID = &value
		}
		operations = append(operations, &operation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return operations, nil
}
