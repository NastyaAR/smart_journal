package repositories

import (
	"context"
	"fmt"

	"blockchain_project/internal/models"

	"github.com/jackc/pgx/v5"
)

type TeacherRepository struct {
	*PostgresRepository
}

func NewTeacherRepository(repo *PostgresRepository) *TeacherRepository {
	return &TeacherRepository{PostgresRepository: repo}
}

func (r *TeacherRepository) CreateTeacher(ctx context.Context, teacher *models.Teacher, userID int) error {
	query := `INSERT INTO teachers (name, email, user_id) VALUES ($1, $2, $3) RETURNING id`
	return r.pool.QueryRow(ctx, query, teacher.Name, teacher.Email, userID).Scan(&teacher.ID)
}

func (r *TeacherRepository) GetTeacherByID(ctx context.Context, id int) (*models.Teacher, error) {
	var teacher models.Teacher
	query := `SELECT t.id, t.name, t.email, t.user_id FROM teachers t WHERE t.id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&teacher.ID, &teacher.Name, &teacher.Email, &teacher.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("teacher not found")
		}
		return nil, fmt.Errorf("failed to get teacher: %w", err)
	}
	return &teacher, nil
}

func (r *TeacherRepository) GetTeacherByUserID(ctx context.Context, userID int) (*models.Teacher, error) {
	var teacher models.Teacher
	query := `SELECT t.id, t.name, t.email, t.user_id FROM teachers t WHERE t.user_id = $1`
	err := r.pool.QueryRow(ctx, query, userID).Scan(&teacher.ID, &teacher.Name, &teacher.Email, &teacher.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("teacher not found")
		}
		return nil, fmt.Errorf("failed to get teacher by user id: %w", err)
	}
	return &teacher, nil
}

func (r *TeacherRepository) GetTeacherByEmail(ctx context.Context, email string) (*models.Teacher, error) {
	var teacher models.Teacher
	query := `SELECT t.id, t.name, t.email, t.user_id FROM teachers t WHERE t.email = $1`
	err := r.pool.QueryRow(ctx, query, email).Scan(&teacher.ID, &teacher.Name, &teacher.Email, &teacher.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("teacher not found")
		}
		return nil, fmt.Errorf("failed to get teacher by email: %w", err)
	}
	return &teacher, nil
}

func (r *TeacherRepository) UpdateTeacher(ctx context.Context, teacher *models.Teacher) error {
	query := `UPDATE teachers SET name = $1, email = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, teacher.Name, teacher.Email, teacher.ID)
	if err != nil {
		return fmt.Errorf("failed to update teacher: %w", err)
	}
	return nil
}

func (r *TeacherRepository) AssignGroup(ctx context.Context, teacherID, groupID int) error {
	query := `INSERT INTO teacher_groups (teacher_id, group_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, query, teacherID, groupID)
	if err != nil {
		return fmt.Errorf("failed to assign group to teacher: %w", err)
	}
	return nil
}

func (r *TeacherRepository) AssignSubject(ctx context.Context, teacherID, subjectID int) error {
	query := `INSERT INTO teacher_subjects (teacher_id, subject_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, query, teacherID, subjectID)
	if err != nil {
		return fmt.Errorf("failed to assign subject to teacher: %w", err)
	}
	return nil
}

func (r *TeacherRepository) HasGroup(ctx context.Context, teacherID, groupID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM teacher_groups WHERE teacher_id = $1 AND group_id = $2)`
	if err := r.pool.QueryRow(ctx, query, teacherID, groupID).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check teacher group access: %w", err)
	}
	return exists, nil
}

func (r *TeacherRepository) HasSubject(ctx context.Context, teacherID, subjectID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM teacher_subjects WHERE teacher_id = $1 AND subject_id = $2)`
	if err := r.pool.QueryRow(ctx, query, teacherID, subjectID).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check teacher subject access: %w", err)
	}
	return exists, nil
}
