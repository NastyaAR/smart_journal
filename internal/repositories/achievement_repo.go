package repositories

import (
	"context"
	"fmt"

	"blockchain_project/internal/models"
)

type AchievementRepository struct {
	*PostgresRepository
}

func NewAchievementRepository(repo *PostgresRepository) *AchievementRepository {
	return &AchievementRepository{PostgresRepository: repo}
}

func (r *AchievementRepository) CreateAchievement(ctx context.Context, achievement *models.Achievement) error {
	var confirmedByTeacherID any
	if achievement.ConfirmedByTeacherID != 0 {
		confirmedByTeacherID = achievement.ConfirmedByTeacherID
	}
	if achievement.Status == "" {
		achievement.Status = "pending"
	}
	achievement.Confirmed = achievement.Status == "confirmed"
	query := `INSERT INTO achievements (student_id, title, description, confirmed, confirmed_by_teacher_id, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	return r.pool.QueryRow(ctx, query, achievement.StudentID, achievement.Title, achievement.Description, achievement.Confirmed, confirmedByTeacherID, achievement.Status).Scan(&achievement.ID)
}

func (r *AchievementRepository) GetAchievementsByStudentID(ctx context.Context, studentID int) ([]*models.Achievement, error) {
	var achievements []*models.Achievement
	query := `SELECT id, student_id, title, description, status, confirmed, COALESCE(confirmed_by_teacher_id, 0) FROM achievements WHERE student_id = $1 ORDER BY id`
	rows, err := r.pool.Query(ctx, query, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievements: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var achievement models.Achievement
		if err := rows.Scan(&achievement.ID, &achievement.StudentID, &achievement.Title, &achievement.Description, &achievement.Status, &achievement.Confirmed, &achievement.ConfirmedByTeacherID); err != nil {
			return nil, fmt.Errorf("failed to scan achievement: %w", err)
		}
		achievements = append(achievements, &achievement)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return achievements, nil
}

func (r *AchievementRepository) GetAchievementByID(ctx context.Context, id int) (*models.Achievement, error) {
	var achievement models.Achievement
	query := `SELECT id, student_id, title, description, status, confirmed, COALESCE(confirmed_by_teacher_id, 0) FROM achievements WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&achievement.ID, &achievement.StudentID, &achievement.Title, &achievement.Description, &achievement.Status, &achievement.Confirmed, &achievement.ConfirmedByTeacherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievement: %w", err)
	}
	return &achievement, nil
}

func (r *AchievementRepository) ConfirmAchievement(ctx context.Context, id int, teacherID int) error {
	query := `UPDATE achievements SET confirmed = true, status = 'confirmed', confirmed_by_teacher_id = $2 WHERE id = $1 AND status = 'pending'`
	tag, err := r.pool.Exec(ctx, query, id, teacherID)
	if err != nil {
		return fmt.Errorf("failed to confirm achievement: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("achievement not found or already processed")
	}
	return nil
}

func (r *AchievementRepository) DenyAchievement(ctx context.Context, id, teacherID int) error {
	query := `UPDATE achievements SET confirmed_by_teacher_id = $1, confirmed = false, status = 'denied' WHERE id = $2 AND status = 'pending'`
	tag, err := r.pool.Exec(ctx, query, teacherID, id)
	if err != nil {
		return fmt.Errorf("failed to deny achievement: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("achievement not found or already processed")
	}
	return nil
}

func (r *AchievementRepository) GetPendingAchievements(ctx context.Context, teacherID int) ([]*models.Achievement, error) {
	query := `
		SELECT a.id, a.student_id, a.title, a.description, a.status, a.confirmed, COALESCE(a.confirmed_by_teacher_id, 0)
		FROM achievements a
		JOIN students s ON s.id = a.student_id
		JOIN teacher_groups tg ON tg.group_id = s.group_id
		WHERE a.status = 'pending' AND tg.teacher_id = $1
		ORDER BY a.id`
	rows, err := r.pool.Query(ctx, query, teacherID)
	if err != nil {
		return nil, fmt.Errorf("failed to query achievements: %w", err)
	}
	defer rows.Close()

	var achievements []*models.Achievement
	for rows.Next() {
		var a models.Achievement
		err := rows.Scan(&a.ID, &a.StudentID, &a.Title, &a.Description, &a.Status, &a.Confirmed, &a.ConfirmedByTeacherID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan achievement: %w", err)
		}
		achievements = append(achievements, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return achievements, nil
}
