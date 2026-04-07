package service

import (
	"context"
	"fmt"

	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/repository"
)

type BillingService interface {
	Create(ctx context.Context, accountID, authUserID string, req *models.CreateBillingRequest) (*models.AccountBilling, error)
	GetAllByAccount(ctx context.Context, accountID, authUserID string, p models.PaginationParams) ([]models.AccountBilling, int, error)
	GetByID(ctx context.Context, id, authUserID string) (*models.AccountBilling, error)
	Update(ctx context.Context, id, authUserID string, req *models.UpdateBillingRequest) (*models.AccountBilling, error)
}

type billingService struct {
	billings repository.BillingRepository
	accounts repository.AccountRepository
}

func NewBillingService(billings repository.BillingRepository, accounts repository.AccountRepository) BillingService {
	return &billingService{billings: billings, accounts: accounts}
}

// nextPeriod calcula el período YYYYMM siguiente dado un período actual.
func nextPeriod(period int) int {
	year := period / 100
	month := period % 100
	if month == 12 {
		return (year+1)*100 + 1
	}
	return period + 1
}

func (s *billingService) Create(ctx context.Context, accountID, authUserID string, req *models.CreateBillingRequest) (*models.AccountBilling, error) {
	year := req.Period / 100
	month := req.Period % 100
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("period inválido: el mes debe estar entre 01 y 12 (formato YYYYMM, ej: 202603)")
	}
	if year < 2000 {
		return nil, fmt.Errorf("period inválido: año mínimo 2000")
	}
	if req.AmountBilled <= 0 {
		return nil, fmt.Errorf("amount_billed debe ser mayor a 0")
	}

	account, err := s.accounts.GetByID(ctx, accountID, authUserID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("not found")
	}

	// Si auto_accumulate y hay una factura impaga, crear carry-over al siguiente período
	if account.AutoAccumulate {
		unpaid, err := s.billings.GetUnpaidByAccount(ctx, accountID)
		if err != nil {
			return nil, err
		}
		if unpaid != nil {
			_, err := s.billings.CreateCarryOver(ctx, accountID, nextPeriod(unpaid.Period), unpaid.AmountBilled, unpaid.ID)
			if err != nil {
				return nil, fmt.Errorf("carry over billing: %w", err)
			}
		}
	}

	return s.billings.Create(ctx, accountID, req)
}

func (s *billingService) GetAllByAccount(ctx context.Context, accountID, authUserID string, p models.PaginationParams) ([]models.AccountBilling, int, error) {
	return s.billings.GetAllByAccount(ctx, accountID, authUserID, p)
}

func (s *billingService) GetByID(ctx context.Context, id, authUserID string) (*models.AccountBilling, error) {
	return s.billings.GetByID(ctx, id, authUserID)
}

func (s *billingService) Update(ctx context.Context, id, authUserID string, req *models.UpdateBillingRequest) (*models.AccountBilling, error) {
	billing, err := s.billings.Update(ctx, id, authUserID, req)
	if err != nil {
		return nil, err
	}
	if billing == nil {
		return nil, fmt.Errorf("not found")
	}

	// Auto-marcar como pagada si amount_paid >= amount_billed
	if !billing.IsPaid && billing.AmountPaid >= billing.AmountBilled {
		if err := s.billings.MarkPaid(ctx, billing.ID); err != nil {
			return nil, err
		}
		billing.IsPaid = true
	}

	return billing, nil
}
