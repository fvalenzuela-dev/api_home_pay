package service

import (
	"context"
	"fmt"

	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/repository"
	"github.com/jackc/pgx/v5"
)

type AccountService interface {
	Create(ctx context.Context, companyID, authUserID string, req *models.CreateAccountRequest) (*models.Account, error)
	GetByID(ctx context.Context, id, authUserID string) (*models.Account, error)
	GetAllByCompany(ctx context.Context, companyID, authUserID string) ([]models.Account, error)
	Update(ctx context.Context, id, authUserID string, req *models.UpdateAccountRequest) (*models.Account, error)
	Delete(ctx context.Context, id, authUserID string) error
}

type accountService struct {
	accounts repository.AccountRepository
	billings repository.BillingRepository
}

func NewAccountService(accounts repository.AccountRepository, billings repository.BillingRepository) AccountService {
	return &accountService{accounts: accounts, billings: billings}
}

func (s *accountService) Create(ctx context.Context, companyID, authUserID string, req *models.CreateAccountRequest) (*models.Account, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.BillingDay < 1 || req.BillingDay > 31 {
		return nil, fmt.Errorf("billing_day must be between 1 and 31")
	}
	return s.accounts.Create(ctx, companyID, authUserID, req)
}

func (s *accountService) GetByID(ctx context.Context, id, authUserID string) (*models.Account, error) {
	return s.accounts.GetByID(ctx, id, authUserID)
}

func (s *accountService) GetAllByCompany(ctx context.Context, companyID, authUserID string) ([]models.Account, error) {
	return s.accounts.GetAllByCompany(ctx, companyID, authUserID)
}

func (s *accountService) Update(ctx context.Context, id, authUserID string, req *models.UpdateAccountRequest) (*models.Account, error) {
	if req.BillingDay != nil && (*req.BillingDay < 1 || *req.BillingDay > 31) {
		return nil, fmt.Errorf("billing_day must be between 1 and 31")
	}
	return s.accounts.Update(ctx, id, authUserID, req)
}

func (s *accountService) Delete(ctx context.Context, id, authUserID string) error {
	if err := s.billings.SoftDeleteByAccount(ctx, id); err != nil {
		return err
	}
	if err := s.accounts.SoftDelete(ctx, id, authUserID); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("not found")
		}
		return err
	}
	return nil
}
