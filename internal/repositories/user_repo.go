package repositories

import (
	"context"
	"fmt"

	"blockchain_project/internal/models"

	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	pg *PostgresRepository
}

func NewUserRepository(pg *PostgresRepository) *UserRepository {
	return &UserRepository{pg: pg}
}

func (r *UserRepository) Authenticate(ctx context.Context, login, password string) (*models.User, error) {
	query := `SELECT id, login, password, role FROM users WHERE login = $1`

	var user models.User
	err := r.pg.pool.QueryRow(ctx, query, login).Scan(
		&user.ID, &user.Login, &user.Password, &user.Role)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if user.Password != password {
		return nil, fmt.Errorf("invalid credentials")
	}

	user.Password = ""
	return &user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (login, password, role) VALUES ($1, $2, $3) RETURNING id`
	return r.pg.pool.QueryRow(ctx, query, user.Login, user.Password, user.Role).Scan(&user.ID)
}

func (r *UserRepository) DeleteUser(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.pg.pool.Exec(ctx, query, id)
	return err
}
