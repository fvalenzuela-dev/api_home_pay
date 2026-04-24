package service

import (
	"context"
	"fmt"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/repository"
	"github.com/jackc/pgx/v5"
)

type InstallmentService interface {
	Create(ctx context.Context, authUserID string, req *models.CreateInstallmentRequest) (*models.InstallmentPlanWithPayments, error)
	GetAll(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.InstallmentPlanWithPayments, int, error)
	GetByID(ctx context.Context, id, authUserID string) (*models.InstallmentPlanWithPayments, error)
	PayInstallment(ctx context.Context, planID, paymentID, authUserID string) (*models.InstallmentPayment, error)
	Delete(ctx context.Context, id, authUserID string) error
}

type installmentService struct {
	installments repository.InstallmentRepository
}

func NewInstallmentService(installments repository.InstallmentRepository) InstallmentService {
	return &installmentService{installments: installments}
}

func (s *installmentService) Create(ctx context.Context, authUserID string, req *models.CreateInstallmentRequest) (*models.InstallmentPlanWithPayments, error) {
	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if req.TotalAmount <= 0 {
		return nil, fmt.Errorf("total_amount must be greater than 0")
	}
	if req.TotalInstallments <= 0 {
		return nil, fmt.Errorf("total_installments must be greater than 0")
	}
	if req.StartDate == "" {
		return nil, fmt.Errorf("start_date is required")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format, expected YYYY-MM-DD")
	}

	plan, err := s.installments.CreatePlan(ctx, authUserID, &models.InstallmentPlan{
		Description:       req.Description,
		TotalAmount:       req.TotalAmount,
		TotalInstallments: req.TotalInstallments,
		StartDate:         startDate,
	})
	if err != nil {
		return nil, err
	}

	installmentAmount := req.TotalAmount / float64(req.TotalInstallments)
	payments := make([]models.InstallmentPayment, req.TotalInstallments)
	for i := range payments {
		payments[i] = models.InstallmentPayment{
			PlanID:            plan.ID,
			InstallmentNumber: i + 1,
			Amount:            installmentAmount,
			DueDate:           startDate.AddDate(0, i, 0),
		}
	}

	if err := s.installments.CreatePayments(ctx, payments); err != nil {
		return nil, fmt.Errorf("create payments: %w", err)
	}

	return &models.InstallmentPlanWithPayments{
		InstallmentPlan: *plan,
		Payments:        payments,
	}, nil
}

// allPayments se usa en llamadas internas que necesitan todos los pagos de un plan.
var allPayments = models.PaginationParams{Page: 1, Limit: 10000}

func (s *installmentService) GetByID(ctx context.Context, id, authUserID string) (*models.InstallmentPlanWithPayments, error) {
	plan, err := s.installments.GetPlan(ctx, id, authUserID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, nil
	}
	payments, _, err := s.installments.GetPaymentsByPlan(ctx, plan.ID, allPayments)
	if err != nil {
		return nil, err
	}
	return &models.InstallmentPlanWithPayments{InstallmentPlan: *plan, Payments: payments}, nil
}

func (s *installmentService) GetAll(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.InstallmentPlanWithPayments, int, error) {
	plans, total, err := s.installments.GetAllPlans(ctx, authUserID, p)
	if err != nil {
		return nil, 0, err
	}

	result := make([]models.InstallmentPlanWithPayments, len(plans))
	for i, plan := range plans {
		payments, _, err := s.installments.GetPaymentsByPlan(ctx, plan.ID, allPayments)
		if err != nil {
			return nil, 0, err
		}
		result[i] = models.InstallmentPlanWithPayments{
			InstallmentPlan: plan,
			Payments:        payments,
		}
	}
	return result, total, nil
}

func (s *installmentService) PayInstallment(ctx context.Context, planID, paymentID, authUserID string) (*models.InstallmentPayment, error) {
	plan, err := s.installments.GetPlan(ctx, planID, authUserID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, fmt.Errorf("not found")
	}

	payment, err := s.installments.UpdatePayment(ctx, planID, paymentID, authUserID)
	if err != nil {
		return nil, err
	}
	if payment == nil {
		return nil, fmt.Errorf("not found or already paid")
	}

	if err := s.installments.IncrementPaid(ctx, planID, plan.TotalInstallments); err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *installmentService) Delete(ctx context.Context, id, authUserID string) error {
	if err := s.installments.SoftDeletePlan(ctx, id, authUserID); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("not found")
		}
		return err
	}
	return nil
}
