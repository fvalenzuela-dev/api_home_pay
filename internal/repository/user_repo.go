package repository

import (
	"context"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Upsert(ctx context.Context, user *models.User) error
	SoftDelete(ctx context.Context, authUserID string) error
}

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Upsert(ctx context.Context, user *models.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO homepay.users (auth_user_id, email, full_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (auth_user_id) DO UPDATE
		SET email      = EXCLUDED.email,
		    full_name  = EXCLUDED.full_name,
		    updated_at = NOW()
		WHERE homepay.users.deleted_at IS NULL
	`, user.AuthUserID, user.Email, user.FullName)
	return err
}

func (r *userRepo) SoftDelete(ctx context.Context, authUserID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE homepay.users
		SET deleted_at = NOW()
		WHERE auth_user_id = $1 AND deleted_at IS NULL
	`, authUserID)
	return err
}
