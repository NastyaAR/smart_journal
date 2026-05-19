package repositories

import (
	"context"
	"fmt"

	"blockchain_project/internal/models"

	"github.com/jackc/pgx/v5"
)

type GroupRepository struct {
	*PostgresRepository
}

func NewGroupRepository(repo *PostgresRepository) *GroupRepository {
	return &GroupRepository{PostgresRepository: repo}
}

func (r *GroupRepository) CreateGroup(ctx context.Context, group *models.Group) error {
	query := `INSERT INTO groups (name) VALUES ($1) RETURNING id`
	return r.pool.QueryRow(ctx, query, group.Name).Scan(&group.ID)
}

func (r *GroupRepository) GetGroupByID(ctx context.Context, id int) (*models.Group, error) {
	var group models.Group
	query := `SELECT id, name FROM groups WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&group.ID, &group.Name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("group not found")
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	return &group, nil
}

func (r *GroupRepository) GetAllGroups(ctx context.Context) ([]*models.Group, error) {
	query := `SELECT id, name FROM groups ORDER BY name`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}
	defer rows.Close()

	var groups []*models.Group
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.ID, &group.Name); err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, &group)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return groups, nil
}

func (r *GroupRepository) GetGroupsByTeacherID(ctx context.Context, teacherID int) ([]*models.Group, error) {
	query := `
		SELECT g.id, g.name
		FROM groups g
		JOIN teacher_groups tg ON tg.group_id = g.id
		WHERE tg.teacher_id = $1
		ORDER BY g.name`
	rows, err := r.pool.Query(ctx, query, teacherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teacher groups: %w", err)
	}
	defer rows.Close()

	var groups []*models.Group
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.ID, &group.Name); err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, &group)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return groups, nil
}
