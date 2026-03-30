package service

import (
	"context"
	"fmt"

	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/repository"
	"github.com/jackc/pgx/v5"
)

type CompanyService interface {
	Create(ctx context.Context, authUserID string, req *models.CreateCompanyRequest) (*models.Company, error)
	GetByID(ctx context.Context, id, authUserID string) (*models.Company, error)
	GetAll(ctx context.Context, authUserID string) ([]models.Company, error)
	Update(ctx context.Context, id, authUserID string, req *models.UpdateCompanyRequest) (*models.Company, error)
	Delete(ctx context.Context, id, authUserID string) error
}

type companyService struct {
	companies repository.CompanyRepository
	accounts  repository.AccountRepository
	billings  repository.BillingRepository
}

func NewCompanyService(companies repository.CompanyRepository, accounts repository.AccountRepository, billings repository.BillingRepository) CompanyService {
	return &companyService{
		companies: companies,
		accounts:  accounts,
		billings:  billings,
	}
}

func (s *companyService) Create(ctx context.Context, authUserID string, req *models.CreateCompanyRequest) (*models.Company, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	return s.companies.Create(ctx, authUserID, req)
}

func (s *companyService) GetByID(ctx context.Context, id, authUserID string) (*models.Company, error) {
	return s.companies.GetByID(ctx, id, authUserID)
}

func (s *companyService) GetAll(ctx context.Context, authUserID string) ([]models.Company, error) {
	return s.companies.GetAll(ctx, authUserID)
}

func (s *companyService) Update(ctx context.Context, id, authUserID string, req *models.UpdateCompanyRequest) (*models.Company, error) {
	return s.companies.Update(ctx, id, authUserID, req)
}

func (s *companyService) Delete(ctx context.Context, id, authUserID string) error {
	accountIDs, err := s.accounts.GetActiveIDsByCompany(ctx, id)
	if err != nil {
		return err
	}
	for _, accountID := range accountIDs {
		if err := s.billings.SoftDeleteByAccount(ctx, accountID); err != nil {
			return err
		}
	}
	if err := s.accounts.SoftDeleteByCompany(ctx, id); err != nil {
		return err
	}
	if err := s.companies.SoftDelete(ctx, id, authUserID); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("not found")
		}
		return err
	}
	return nil
}
