package service

import (
	"context"
	"fmt"

	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/repository"
	"github.com/jackc/pgx/v5"
)

type AccountGroupService interface {
	Create(ctx context.Context, authUserID string, req *models.CreateAccountGroupRequest) (*models.AccountGroup, error)
	GetByID(ctx context.Context, id, authUserID string) (*models.AccountGroup, error)
	GetAll(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.AccountGroup, int, error)
	Update(ctx context.Context, id, authUserID string, req *models.UpdateAccountGroupRequest) (*models.AccountGroup, error)
	Delete(ctx context.Context, id, authUserID string) error
}

type accountGroupService struct {
	repo repository.AccountGroupRepository
}

func NewAccountGroupService(repo repository.AccountGroupRepository) AccountGroupService {
	return &accountGroupService{repo: repo}
}

func (s *accountGroupService) Create(ctx context.Context, authUserID string, req *models.CreateAccountGroupRequest) (*models.AccountGroup, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	g, err := s.repo.Create(ctx, authUserID, req)
	if err != nil {
		if err == repository.ErrDuplicateName {
			return nil, fmt.Errorf("ya existe un grupo con ese nombre")
		}
		return nil, err
	}
	return g, nil
}

func (s *accountGroupService) GetByID(ctx context.Context, id, authUserID string) (*models.AccountGroup, error) {
	return s.repo.GetByID(ctx, id, authUserID)
}

func (s *accountGroupService) GetAll(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.AccountGroup, int, error) {
	return s.repo.GetAll(ctx, authUserID, p)
}

func (s *accountGroupService) Update(ctx context.Context, id, authUserID string, req *models.UpdateAccountGroupRequest) (*models.AccountGroup, error) {
	if req.Name != nil && *req.Name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	g, err := s.repo.Update(ctx, id, authUserID, req)
	if err != nil {
		if err == repository.ErrDuplicateName {
			return nil, fmt.Errorf("ya existe un grupo con ese nombre")
		}
		return nil, err
	}
	return g, nil
}

func (s *accountGroupService) Delete(ctx context.Context, id, authUserID string) error {
	if err := s.repo.SoftDelete(ctx, id, authUserID); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("not found")
		}
		return err
	}
	return nil
}
