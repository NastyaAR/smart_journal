package repositories

import (
	"context"
	"fmt"

	"blockchain_project/internal/models"

	"github.com/jackc/pgx/v5"
)

type MerchRepository struct {
	*PostgresRepository
}

func NewMerchRepository(repo *PostgresRepository) *MerchRepository {
	return &MerchRepository{PostgresRepository: repo}
}

func (r *MerchRepository) CreateMerch(ctx context.Context, merch *models.Merch) error {
	query := `INSERT INTO merch (title, description, price) VALUES ($1, $2, $3) RETURNING id`
	return r.pool.QueryRow(ctx, query, merch.Title, merch.Description, merch.Price).Scan(&merch.ID)
}

func (r *MerchRepository) GetMerchByID(ctx context.Context, id int) (*models.Merch, error) {
	var merch models.Merch
	query := `SELECT id, title, description, price FROM merch WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&merch.ID, &merch.Title, &merch.Description, &merch.Price)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("merch not found")
		}
		return nil, fmt.Errorf("failed to get merch: %w", err)
	}
	return &merch, nil
}

func (r *MerchRepository) GetAllMerch(ctx context.Context) ([]*models.Merch, error) {
	var merchList []*models.Merch
	query := `SELECT id, title, description, price FROM merch`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get merch list: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var merch models.Merch
		if err := rows.Scan(&merch.ID, &merch.Title, &merch.Description, &merch.Price); err != nil {
			return nil, fmt.Errorf("failed to scan merch: %w", err)
		}
		merchList = append(merchList, &merch)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return merchList, nil
}

func (r *MerchRepository) GetPurchasesByStudentID(ctx context.Context, studentID int) ([]*models.Purchase, error) {
	query := `
		SELECT p.id, p.student_id, p.merch_id, m.title, p.price, p.created_at
		FROM purchases p
		JOIN merch m ON m.id = p.merch_id
		WHERE p.student_id = $1
		ORDER BY p.created_at DESC, p.id DESC`
	rows, err := r.pool.Query(ctx, query, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchases: %w", err)
	}
	defer rows.Close()

	var purchases []*models.Purchase
	for rows.Next() {
		var purchase models.Purchase
		if err := rows.Scan(&purchase.ID, &purchase.StudentID, &purchase.MerchID, &purchase.Title, &purchase.Price, &purchase.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan purchase: %w", err)
		}
		purchases = append(purchases, &purchase)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return purchases, nil
}
