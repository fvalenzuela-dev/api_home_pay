package service

import (
	"context"

	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/repository"
)

type DashboardService interface {
	GetSummary(ctx context.Context, authUserID string, month, year int) (*DashboardSummary, error)
}

type DashboardSummary struct {
	Month              int                `json:"month"`
	Year               int                `json:"year"`
	TotalBilled        float64            `json:"total_billed"`
	TotalPaid          float64            `json:"total_paid"`
	TotalPending       float64            `json:"total_pending"`
	ExpensesByCompany map[string]float64 `json:"expenses_by_company"`
	TotalExpenses      float64            `json:"total_expenses"`
	TotalInstallments  float64            `json:"total_installments"`
	PendingCommitments []PendingCommitment `json:"pending_commitments"`
}

type PendingCommitment struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

type dashboardService struct {
	billings     repository.BillingRepository
	expenses     repository.ExpenseRepository
	installments repository.InstallmentRepository
}

func NewDashboardService(
	billings repository.BillingRepository,
	expenses repository.ExpenseRepository,
	installments repository.InstallmentRepository,
) DashboardService {
	return &dashboardService{
		billings:     billings,
		expenses:     expenses,
		installments: installments,
	}
}

func (s *dashboardService) GetSummary(ctx context.Context, authUserID string, month, year int) (*DashboardSummary, error) {
	summary := &DashboardSummary{
		Month:             month,
		Year:              year,
		ExpensesByCompany: make(map[string]float64),
	}

	period := year*100 + month

	allPages := models.PaginationParams{Page: 1, Limit: 10000}
	billings, _, err := s.billings.GetAllByPeriod(ctx, authUserID, period, nil, allPages)
	if err != nil {
		return nil, err
	}
	for _, b := range billings {
		summary.TotalBilled += b.AmountBilled
		summary.TotalPaid += b.AmountPaid
		if !b.IsPaid {
			pending := b.AmountBilled - b.AmountPaid
			summary.TotalPending += pending
			summary.PendingCommitments = append(summary.PendingCommitments, PendingCommitment{
				Type:        "billing",
				Description: b.AccountID,
				Amount:      pending,
			})
		}
	}

	expenses, _, err := s.expenses.GetAll(ctx, authUserID, models.ExpenseFilters{Month: &month, Year: &year}, allPages)
	if err != nil {
		return nil, err
	}
	for _, e := range expenses {
		summary.TotalExpenses += e.Amount
		if e.CompanyID != nil {
			summary.ExpensesByCompany[*e.CompanyID] += e.Amount
		}
	}

	installmentPayments, err := s.installments.GetActivePaymentsByMonth(ctx, authUserID, month, year)
	if err != nil {
		return nil, err
	}
	for _, p := range installmentPayments {
		summary.TotalInstallments += p.Amount
		if !p.IsPaid {
			summary.PendingCommitments = append(summary.PendingCommitments, PendingCommitment{
				Type:        "installment",
				Description: p.PlanID,
				Amount:      p.Amount,
			})
		}
	}

	return summary, nil
}
