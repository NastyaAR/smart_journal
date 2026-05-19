package repositories

import (
	"context"
	"fmt"

	"blockchain_project/internal/models"

	"github.com/jackc/pgx/v5"
)

type StudentRepository struct {
	*PostgresRepository
}

func NewStudentRepository(repo *PostgresRepository) *StudentRepository {
	return &StudentRepository{PostgresRepository: repo}
}

func (r *StudentRepository) CreateStudent(ctx context.Context, student *models.Student, userID int) error {
	var groupID any
	if student.GroupID != 0 {
		groupID = student.GroupID
	}
	query := `INSERT INTO students (name, email, group_id, user_id) VALUES ($1, $2, $3, $4) RETURNING id`
	return r.pool.QueryRow(ctx, query, student.Name, student.Email, groupID, userID).Scan(&student.ID)
}

func (r *StudentRepository) DeleteStudent(ctx context.Context, id int) error {
	query := `DELETE FROM students WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete student: %w", err)
	}
	return nil
}

func (r *StudentRepository) GetStudentByID(ctx context.Context, id int) (*models.Student, error) {
	var student models.Student
	query := `SELECT s.id, s.name, s.email, COALESCE(s.group_id, 0), s.tokens, s.user_id FROM students s WHERE s.id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&student.ID, &student.Name, &student.Email, &student.GroupID, &student.Tokens, &student.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("student not found")
		}
		return nil, fmt.Errorf("failed to get student: %w", err)
	}
	return &student, nil
}

func (r *StudentRepository) GetStudentByUserID(ctx context.Context, userID int) (*models.Student, error) {
	var student models.Student
	query := `SELECT s.id, s.name, s.email, COALESCE(s.group_id, 0), s.tokens, s.user_id FROM students s WHERE s.user_id = $1`
	err := r.pool.QueryRow(ctx, query, userID).Scan(&student.ID, &student.Name, &student.Email, &student.GroupID, &student.Tokens, &student.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("student not found")
		}
		return nil, fmt.Errorf("failed to get student by user id: %w", err)
	}
	return &student, nil
}

func (r *StudentRepository) GetStudentByEmail(ctx context.Context, email string) (*models.Student, error) {
	var student models.Student
	query := `SELECT s.id, s.name, s.email, COALESCE(s.group_id, 0), s.tokens, s.user_id FROM students s WHERE s.email = $1`
	err := r.pool.QueryRow(ctx, query, email).Scan(&student.ID, &student.Name, &student.Email, &student.GroupID, &student.Tokens, &student.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("student not found")
		}
		return nil, fmt.Errorf("failed to get student by email: %w", err)
	}
	return &student, nil
}

func (r *StudentRepository) UpdateStudent(ctx context.Context, student *models.Student) error {
	var groupID any
	if student.GroupID != 0 {
		groupID = student.GroupID
	}
	query := `UPDATE students SET name = $1, email = $2, group_id = $3 WHERE id = $4`
	_, err := r.pool.Exec(ctx, query, student.Name, student.Email, groupID, student.ID)
	if err != nil {
		return fmt.Errorf("failed to update student: %w", err)
	}
	return nil
}

func (r *StudentRepository) UpdateStudentTokens(ctx context.Context, id int, tokens int) error {
	query := `UPDATE students SET tokens = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, tokens, id)
	if err != nil {
		return fmt.Errorf("failed to update student tokens: %w", err)
	}
	return nil
}

func (r *StudentRepository) AddStudentTokens(ctx context.Context, id int, amount int) (int, error) {
	var tokens int
	query := `UPDATE students SET tokens = tokens + $1 WHERE id = $2 RETURNING tokens`
	err := r.pool.QueryRow(ctx, query, amount, id).Scan(&tokens)
	if err != nil {
		return 0, fmt.Errorf("failed to add student tokens: %w", err)
	}
	return tokens, nil
}

func (r *StudentRepository) SpendStudentTokens(ctx context.Context, id int, amount int) (int, error) {
	var tokens int
	query := `UPDATE students SET tokens = tokens - $1 WHERE id = $2 AND tokens >= $1 RETURNING tokens`
	err := r.pool.QueryRow(ctx, query, amount, id).Scan(&tokens)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("insufficient tokens")
		}
		return 0, fmt.Errorf("failed to spend student tokens: %w", err)
	}
	return tokens, nil
}

func (r *StudentRepository) PurchaseMerch(ctx context.Context, studentID, merchID, price int) (int, int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to start purchase transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var tokens int
	updateQuery := `UPDATE students SET tokens = tokens - $1 WHERE id = $2 AND tokens >= $1 RETURNING tokens`
	if err := tx.QueryRow(ctx, updateQuery, price, studentID).Scan(&tokens); err != nil {
		if err == pgx.ErrNoRows {
			return 0, 0, fmt.Errorf("insufficient tokens")
		}
		return 0, 0, fmt.Errorf("failed to spend student tokens: %w", err)
	}

	var purchaseID int
	insertQuery := `INSERT INTO purchases (student_id, merch_id, price) VALUES ($1, $2, $3) RETURNING id`
	if err := tx.QueryRow(ctx, insertQuery, studentID, merchID, price).Scan(&purchaseID); err != nil {
		return 0, 0, fmt.Errorf("failed to create purchase: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("failed to commit purchase: %w", err)
	}

	return tokens, purchaseID, nil
}

func (r *StudentRepository) AddToGroup(ctx context.Context, studentID, groupID int) error {
	query := `UPDATE students SET group_id = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, groupID, studentID)
	return err
}

func (r *StudentRepository) GetStudentsByGroupID(ctx context.Context, groupID int) ([]*models.Student, error) {
	query := `SELECT id, name, email, COALESCE(group_id, 0), tokens, user_id FROM students WHERE group_id = $1 ORDER BY name`
	rows, err := r.pool.Query(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get students by group: %w", err)
	}
	defer rows.Close()

	var students []*models.Student
	for rows.Next() {
		var student models.Student
		if err := rows.Scan(&student.ID, &student.Name, &student.Email, &student.GroupID, &student.Tokens, &student.UserID); err != nil {
			return nil, fmt.Errorf("failed to scan student: %w", err)
		}
		students = append(students, &student)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return students, nil
}
