package repositories

import (
	"context"
	"fmt"

	"blockchain_project/internal/models"
)

type GradeRepository struct {
	*PostgresRepository
}

func NewGradeRepository(repo *PostgresRepository) *GradeRepository {
	return &GradeRepository{PostgresRepository: repo}
}

func (r *GradeRepository) CreateGrade(ctx context.Context, grade *models.Grade) error {
	query := `
		INSERT INTO grades (student_id, subject_id, value, lesson_date)
		VALUES ($1, $2, $3, COALESCE(NULLIF($4, '')::date, CURRENT_DATE))
		RETURNING id, lesson_date::text, created_at`
	if err := r.pool.QueryRow(ctx, query, grade.StudentID, grade.SubjectID, grade.Value, grade.LessonDate).Scan(&grade.ID, &grade.LessonDate, &grade.CreatedAt); err != nil {
		return fmt.Errorf("failed to create grade: %w", err)
	}
	return nil
}

func (r *GradeRepository) GetGradesByStudentID(ctx context.Context, studentID int) ([]*models.Grade, error) {
	var grades []*models.Grade
	query := `SELECT id, student_id, subject_id, value, lesson_date::text, created_at FROM grades WHERE student_id = $1`
	rows, err := r.pool.Query(ctx, query, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var grade models.Grade
		if err := rows.Scan(&grade.ID, &grade.StudentID, &grade.SubjectID, &grade.Value, &grade.LessonDate, &grade.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan grade: %w", err)
		}
		grades = append(grades, &grade)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return grades, nil
}

func (r *GradeRepository) GetGradesByGroupID(ctx context.Context, groupID int) ([]*models.Grade, error) {
	var grades []*models.Grade
	query := `
		SELECT g.id, g.student_id, g.subject_id, g.value, g.lesson_date::text, g.created_at
		FROM grades g
		JOIN students s ON s.id = g.student_id
		WHERE s.group_id = $1
		ORDER BY s.name, g.id`
	rows, err := r.pool.Query(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group grades: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var grade models.Grade
		if err := rows.Scan(&grade.ID, &grade.StudentID, &grade.SubjectID, &grade.Value, &grade.LessonDate, &grade.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan grade: %w", err)
		}
		grades = append(grades, &grade)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return grades, nil
}

func (r *GradeRepository) GetGradeViewsByGroupID(ctx context.Context, groupID int) ([]*models.GradeView, error) {
	query := `
		SELECT g.id, g.student_id, s.name, g.subject_id, sub.name, g.value, g.lesson_date::text, g.created_at
		FROM grades g
		JOIN students s ON s.id = g.student_id
		JOIN subjects sub ON sub.id = g.subject_id
		WHERE s.group_id = $1
		ORDER BY s.name, sub.name, g.lesson_date DESC, g.id DESC`
	rows, err := r.pool.Query(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group grade views: %w", err)
	}
	defer rows.Close()

	var grades []*models.GradeView
	for rows.Next() {
		var grade models.GradeView
		if err := rows.Scan(&grade.ID, &grade.StudentID, &grade.StudentName, &grade.SubjectID, &grade.SubjectName, &grade.Value, &grade.LessonDate, &grade.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan grade view: %w", err)
		}
		grades = append(grades, &grade)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return grades, nil
}

func (r *GradeRepository) GetGradeViewsByStudentID(ctx context.Context, studentID int) ([]*models.GradeView, error) {
	query := `
		SELECT g.id, g.student_id, s.name, g.subject_id, sub.name, g.value, g.lesson_date::text, g.created_at
		FROM grades g
		JOIN students s ON s.id = g.student_id
		JOIN subjects sub ON sub.id = g.subject_id
		WHERE g.student_id = $1
		ORDER BY g.lesson_date DESC, sub.name, g.id DESC`
	rows, err := r.pool.Query(ctx, query, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student grade views: %w", err)
	}
	defer rows.Close()

	var grades []*models.GradeView
	for rows.Next() {
		var grade models.GradeView
		if err := rows.Scan(&grade.ID, &grade.StudentID, &grade.StudentName, &grade.SubjectID, &grade.SubjectName, &grade.Value, &grade.LessonDate, &grade.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan grade view: %w", err)
		}
		grades = append(grades, &grade)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return grades, nil
}
