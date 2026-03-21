package repository

import (
	"context"
	"database/sql"
)

type BaseRepository interface {
	WithUserID(userID string) BaseRepository
	GetDB() *sql.DB
	GetUserID() string
	GetContext() context.Context
}

type baseRepository struct {
	db     *sql.DB
	userID string
	ctx    context.Context
}

func NewBaseRepository(db *sql.DB) BaseRepository {
	return &baseRepository{
		db:  db,
		ctx: context.Background(),
	}
}

func (r *baseRepository) WithUserID(userID string) BaseRepository {
	return &baseRepository{
		db:     r.db,
		userID: userID,
		ctx:    r.ctx,
	}
}

func (r *baseRepository) GetDB() *sql.DB {
	return r.db
}

func (r *baseRepository) GetUserID() string {
	return r.userID
}

func (r *baseRepository) GetContext() context.Context {
	return r.ctx
}
