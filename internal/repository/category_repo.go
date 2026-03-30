package repository

import (
	"context"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository interface {
	GetAll(ctx context.Context, authUserID string) ([]models.Category, error)
	GetByID(ctx context.Context, id int, authUserID string) (*models.Category, error)
	Create(ctx context.Context, authUserID string, req *models.CreateCategoryRequest) (*models.Category, error)
	Update(ctx context.Context, id int, authUserID string, req *models.UpdateCategoryRequest) (*models.Category, error)
	Delete(ctx context.Context, id int, authUserID string) error
}

type categoryRepo struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) CategoryRepository {
	return &categoryRepo{db: db}
}

const categoryCols = `id, name, auth_user_id, created_at, updated_at, deleted_at`

func scanCategory(row pgx.Row, c *models.Category) error {
	return row.Scan(&c.ID, &c.Name, &c.AuthUserID, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
}

func (r *categoryRepo) GetAll(ctx context.Context, authUserID string) ([]models.Category, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+categoryCols+`
		FROM homepay.categories
		WHERE auth_user_id = $1 AND deleted_at IS NULL
		ORDER BY name
	`, authUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.AuthUserID, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *categoryRepo) GetByID(ctx context.Context, id int, authUserID string) (*models.Category, error) {
	var c models.Category
	err := scanCategory(r.db.QueryRow(ctx, `
		SELECT `+categoryCols+`
		FROM homepay.categories
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
	`, id, authUserID), &c)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepo) Create(ctx context.Context, authUserID string, req *models.CreateCategoryRequest) (*models.Category, error) {
	var c models.Category
	err := scanCategory(r.db.QueryRow(ctx, `
		INSERT INTO homepay.categories (name, auth_user_id)
		VALUES ($1, $2)
		RETURNING `+categoryCols,
		req.Name, authUserID), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepo) Update(ctx context.Context, id int, authUserID string, req *models.UpdateCategoryRequest) (*models.Category, error) {
	var c models.Category
	err := scanCategory(r.db.QueryRow(ctx, `
		UPDATE homepay.categories
		SET name       = COALESCE($3, name),
		    updated_at = NOW()
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
		RETURNING `+categoryCols,
		id, authUserID, req.Name), &c)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepo) Delete(ctx context.Context, id int, authUserID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE homepay.categories
		SET deleted_at = NOW()
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
	`, id, authUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
