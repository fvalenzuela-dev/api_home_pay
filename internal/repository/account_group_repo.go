package repository

import (
	"context"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountGroupRepository interface {
	Create(ctx context.Context, authUserID string, req *models.CreateAccountGroupRequest) (*models.AccountGroup, error)
	GetByID(ctx context.Context, id, authUserID string) (*models.AccountGroup, error)
	GetAll(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.AccountGroup, int, error)
	Update(ctx context.Context, id, authUserID string, req *models.UpdateAccountGroupRequest) (*models.AccountGroup, error)
	SoftDelete(ctx context.Context, id, authUserID string) error
}

type accountGroupRepo struct {
	db *pgxpool.Pool
}

func NewAccountGroupRepository(db *pgxpool.Pool) AccountGroupRepository {
	return &accountGroupRepo{db: db}
}

const accountGroupCols = `id, auth_user_id, name, created_at, deleted_at`

func scanAccountGroup(row pgx.Row, g *models.AccountGroup) error {
	return row.Scan(&g.ID, &g.AuthUserID, &g.Name, &g.CreatedAt, &g.DeletedAt)
}

func (r *accountGroupRepo) Create(ctx context.Context, authUserID string, req *models.CreateAccountGroupRequest) (*models.AccountGroup, error) {
	var g models.AccountGroup
	err := scanAccountGroup(r.db.QueryRow(ctx, `
		INSERT INTO homepay.account_groups (auth_user_id, name)
		VALUES ($1, $2)
		RETURNING `+accountGroupCols,
		authUserID, req.Name), &g)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateName
		}
		return nil, err
	}
	return &g, nil
}

func (r *accountGroupRepo) GetByID(ctx context.Context, id, authUserID string) (*models.AccountGroup, error) {
	var g models.AccountGroup
	err := scanAccountGroup(r.db.QueryRow(ctx, `
		SELECT `+accountGroupCols+`
		FROM homepay.account_groups
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
	`, id, authUserID), &g)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *accountGroupRepo) GetAll(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.AccountGroup, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM homepay.account_groups
		WHERE auth_user_id = $1 AND deleted_at IS NULL
	`, authUserID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+accountGroupCols+`
		FROM homepay.account_groups
		WHERE auth_user_id = $1 AND deleted_at IS NULL
		ORDER BY name
		LIMIT $2 OFFSET $3
	`, authUserID, p.Limit, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var groups []models.AccountGroup
	for rows.Next() {
		var g models.AccountGroup
		if err := rows.Scan(&g.ID, &g.AuthUserID, &g.Name, &g.CreatedAt, &g.DeletedAt); err != nil {
			return nil, 0, err
		}
		groups = append(groups, g)
	}
	return groups, total, rows.Err()
}

func (r *accountGroupRepo) Update(ctx context.Context, id, authUserID string, req *models.UpdateAccountGroupRequest) (*models.AccountGroup, error) {
	var g models.AccountGroup
	err := scanAccountGroup(r.db.QueryRow(ctx, `
		UPDATE homepay.account_groups
		SET name = COALESCE($3, name)
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL
		RETURNING `+accountGroupCols,
		id, authUserID, req.Name), &g)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateName
		}
		return nil, err
	}
	return &g, nil
}

func (r *accountGroupRepo) SoftDelete(ctx context.Context, id, authUserID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE homepay.account_groups
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
