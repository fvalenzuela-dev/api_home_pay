package repository

import (
	"context"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CompanyRepository interface {
	Create(ctx context.Context, authUserID string, req *models.CreateCompanyRequest) (*models.Company, error)
	GetByID(ctx context.Context, id, authUserID string) (*models.Company, error)
	GetAll(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.Company, int, error)
	Update(ctx context.Context, id, authUserID string, req *models.UpdateCompanyRequest) (*models.Company, error)
	SoftDelete(ctx context.Context, id, authUserID string) error
}

type companyRepo struct {
	db *pgxpool.Pool
}

func NewCompanyRepository(db *pgxpool.Pool) CompanyRepository {
	return &companyRepo{db: db}
}

const companyCols = `id, auth_user_id, category_id, name, website, phone, is_active, created_at, deleted_at`

func scanCompany(row pgx.Row, c *models.Company) error {
	return row.Scan(&c.ID, &c.AuthUserID, &c.CategoryID, &c.Name, &c.Website, &c.Phone, &c.IsActive, &c.CreatedAt, &c.DeletedAt)
}

func scanCompanyWithCategory(row pgx.Row, c *models.Company) error {
	return row.Scan(&c.ID, &c.AuthUserID, &c.CategoryID, &c.CategoryName, &c.Name, &c.Website, &c.Phone, &c.IsActive, &c.CreatedAt, &c.DeletedAt)
}

func (r *companyRepo) Create(ctx context.Context, authUserID string, req *models.CreateCompanyRequest) (*models.Company, error) {
	var c models.Company
	err := scanCompany(r.db.QueryRow(ctx, `
		INSERT INTO homepay.companies (auth_user_id, category_id, name, website, phone)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+companyCols,
		authUserID, req.CategoryID, req.Name, req.Website, req.Phone), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *companyRepo) GetByID(ctx context.Context, id, authUserID string) (*models.Company, error) {
	var c models.Company
	err := scanCompany(r.db.QueryRow(ctx, `
		SELECT `+companyCols+`
		FROM homepay.companies
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

func (r *companyRepo) GetAll(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.Company, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM homepay.companies
		WHERE auth_user_id = $1 AND deleted_at IS NULL
	`, authUserID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT co.id, co.auth_user_id, co.category_id, c.name AS category_name,
		       co.name, co.website, co.phone, co.is_active, co.created_at, co.deleted_at
		FROM homepay.companies co
		LEFT JOIN homepay.categories c ON c.id = co.category_id
		WHERE co.auth_user_id = $1 AND co.deleted_at IS NULL
		ORDER BY co.created_at DESC
		LIMIT $2 OFFSET $3
	`, authUserID, p.Limit, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var companies []models.Company
	for rows.Next() {
		var c models.Company
		if err := rows.Scan(&c.ID, &c.AuthUserID, &c.CategoryID, &c.CategoryName, &c.Name, &c.Website, &c.Phone, &c.IsActive, &c.CreatedAt, &c.DeletedAt); err != nil {
			return nil, 0, err
		}
		companies = append(companies, c)
	}
	return companies, total, rows.Err()
}

func (r *companyRepo) Update(ctx context.Context, id, authUserID string, req *models.UpdateCompanyRequest) (*models.Company, error) {
	var c models.Company
	err := scanCompany(r.db.QueryRow(ctx, `
		UPDATE homepay.companies
		SET name        = COALESCE($3, name),
		    category_id = COALESCE($4, category_id),
		    website     = COALESCE($5, website),
		    phone       = COALESCE($6, phone)
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
		RETURNING `+companyCols,
		id, authUserID, req.Name, req.CategoryID, req.Website, req.Phone), &c)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *companyRepo) SoftDelete(ctx context.Context, id, authUserID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE homepay.companies
		SET deleted_at = NOW(), is_active = FALSE
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
