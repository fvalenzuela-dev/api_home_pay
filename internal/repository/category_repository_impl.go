package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
)

type categoryRepository struct {
	db  *sql.DB
	ctx context.Context
}

func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepository{
		db:  db,
		ctx: context.Background(),
	}
}

func (r *categoryRepository) Create(userID string, category *models.Category) error {
	query := `
		INSERT INTO categories (name, created_at)
		VALUES ($1, NOW())
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(r.ctx, query, category.Name).Scan(&category.ID, &category.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

func (r *categoryRepository) GetByID(userID string, id int) (*models.Category, error) {
	query := `
		SELECT id, name, created_at
		FROM categories
		WHERE id = $1
	`

	category := &models.Category{}
	err := r.db.QueryRowContext(r.ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category by ID: %w", err)
	}

	return category, nil
}

func (r *categoryRepository) GetAll(userID string) ([]models.Category, error) {
	query := `
		SELECT id, name, created_at
		FROM categories
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(r.ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}

func (r *categoryRepository) Update(userID string, category *models.Category) error {
	query := `
		UPDATE categories
		SET name = $1
		WHERE id = $2
		RETURNING created_at
	`

	var createdAt string
	err := r.db.QueryRowContext(r.ctx, query, category.Name, category.ID).Scan(&createdAt)
	if err == sql.ErrNoRows {
		return fmt.Errorf("category not found or access denied")
	}
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	category.CreatedAt = createdAt
	return nil
}

func (r *categoryRepository) Delete(userID string, id int) error {
	query := `
		DELETE FROM categories
		WHERE id = $1
	`

	result, err := r.db.ExecContext(r.ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found or access denied")
	}

	return nil
}

func (r *categoryRepository) ExistsByName(userID string, name string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM categories
			WHERE name = $1
		)
	`

	var exists bool
	err := r.db.QueryRowContext(r.ctx, query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check category existence: %w", err)
	}

	return exists, nil
}

func (r *categoryRepository) HasExpenses(id int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM expenses
			WHERE category_id = $1
		)
	`

	var exists bool
	err := r.db.QueryRowContext(r.ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check category expenses: %w", err)
	}

	return exists, nil
}
