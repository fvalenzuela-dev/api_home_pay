package service

import (
	"context"
	"fmt"

	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/repository"
)

type BillingService interface {
	Create(ctx context.Context, accountID, authUserID string, req *models.CreateBillingRequest) (*models.AccountBilling, error)
	GetAllByAccount(ctx context.Context, accountID, authUserID string) ([]models.AccountBilling, error)
	Update(ctx context.Context, id, authUserID string, req *models.UpdateBillingRequest) (*models.AccountBilling, error)
}

type billingService struct {
	billings repository.BillingRepository
	accounts repository.AccountRepository
}

func NewBillingService(billings repository.BillingRepository, accounts repository.AccountRepository) BillingService {
	return &billingService{billings: billings, accounts: accounts}
}

func (s *billingService) Create(ctx context.Context, accountID, authUserID string, req *models.CreateBillingRequest) (*models.AccountBilling, error) {
	if req.Month < 1 || req.Month > 12 {
		return nil, fmt.Errorf("month must be between 1 and 12")
	}
	if req.Year < 2000 {
		return nil, fmt.Errorf("invalid year")
	}
	if req.AmountBilled <= 0 {
		return nil, fmt.Errorf("amount_billed must be greater than 0")
	}

	// Verificar si hay billing impaga con auto_accumulate — si la hay, acumularla
	account, err := s.accounts.GetByID(ctx, accountID, authUserID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("not found")
	}

	if account.AutoAccumulate {
		unpaid, err := s.billings.GetUnpaidByAccount(ctx, accountID)
		if err != nil {
			return nil, err
		}
		if unpaid != nil {
			nextMonth := unpaid.Month + 1
			nextYear := unpaid.Year
			if nextMonth > 12 {
				nextMonth = 1
				nextYear++
			}
			_, err := s.billings.CreateCarryOver(ctx, accountID, nextMonth, nextYear, unpaid.AmountBilled, unpaid.ID)
			if err != nil {
				return nil, fmt.Errorf("carry over billing: %w", err)
			}
		}
	}

	return s.billings.Create(ctx, accountID, req)
}

func (s *billingService) GetAllByAccount(ctx context.Context, accountID, authUserID string) ([]models.AccountBilling, error) {
	return s.billings.GetAllByAccount(ctx, accountID, authUserID)
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
