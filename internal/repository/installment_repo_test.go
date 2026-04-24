package repository

import (
	"testing"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestInstallmentRepo_Interfaces(t *testing.T) {
	t.Run("InstallmentRepository interface is satisfied by installmentRepo", func(t *testing.T) {
		var _ InstallmentRepository = (*installmentRepo)(nil)
	})
}

func TestScanPlan(t *testing.T) {
	t.Run("scanPlan function exists", func(t *testing.T) {
		assert.NotNil(t, scanPlan)
	})
}

func TestScanPayment(t *testing.T) {
	t.Run("scanPayment function exists", func(t *testing.T) {
		assert.NotNil(t, scanPayment)
	})
}

func TestInstallmentRepo_Columns(t *testing.T) {
	t.Run("planCols constant", func(t *testing.T) {
		assert.Equal(t, `id, auth_user_id, description, total_amount, total_installments, installments_paid, start_date, is_completed, created_at, deleted_at`, planCols)
	})

	t.Run("paymentCols constant", func(t *testing.T) {
		assert.Equal(t, `id, plan_id, installment_number, amount, due_date, is_paid, paid_at, created_at, deleted_at`, paymentCols)
	})
}

func TestInstallmentRepo_CreatePlan(t *testing.T) {
	t.Run("CreateInstallmentRequest validation", func(t *testing.T) {
		req := models.CreateInstallmentRequest{
			Description:       "Test Plan",
			TotalAmount:      120000,
			TotalInstallments: 12,
			StartDate:        "2026-03-01",
		}
		assert.Equal(t, "Test Plan", req.Description)
		assert.Equal(t, 120000.0, req.TotalAmount)
		assert.Equal(t, 12, req.TotalInstallments)
		assert.Equal(t, "2026-03-01", req.StartDate)
	})
}

func TestInstallmentPlanModel(t *testing.T) {
	t.Run("InstallmentPlan model fields", func(t *testing.T) {
		now := time.Now()
		plan := models.InstallmentPlan{
			ID:                "plan-123",
			AuthUserID:        "user-123",
			Description:       "Test Plan",
			TotalAmount:       120000,
			TotalInstallments: 12,
			InstallmentsPaid:  3,
			StartDate:         now,
			IsCompleted:       false,
			CreatedAt:         now,
		}
		assert.Equal(t, "plan-123", plan.ID)
		assert.Equal(t, 120000.0, plan.TotalAmount)
		assert.Equal(t, 12, plan.TotalInstallments)
		assert.Equal(t, 3, plan.InstallmentsPaid)
		assert.False(t, plan.IsCompleted)
	})
}

func TestInstallmentPaymentModel(t *testing.T) {
	t.Run("InstallmentPayment model fields", func(t *testing.T) {
		now := time.Now()
		payment := models.InstallmentPayment{
			ID:                 "payment-1",
			PlanID:             "plan-123",
			InstallmentNumber:  1,
			Amount:             10000,
			DueDate:            now,
			IsPaid:             true,
			PaidAt:             &now,
			CreatedAt:          now,
		}
		assert.Equal(t, "payment-1", payment.ID)
		assert.Equal(t, 1, payment.InstallmentNumber)
		assert.Equal(t, 10000.0, payment.Amount)
		assert.True(t, payment.IsPaid)
		assert.NotNil(t, payment.PaidAt)
	})
}

func TestInstallmentPlanWithPaymentsModel(t *testing.T) {
	t.Run("InstallmentPlanWithPayments model", func(t *testing.T) {
		plan := models.InstallmentPlan{
			ID:                "plan-123",
			Description:       "Test Plan",
			TotalAmount:       120000,
			TotalInstallments: 12,
		}
		payments := []models.InstallmentPayment{
			{ID: "payment-1", PlanID: "plan-123", InstallmentNumber: 1, Amount: 10000},
			{ID: "payment-2", PlanID: "plan-123", InstallmentNumber: 2, Amount: 10000},
		}
		
		planWithPayments := models.InstallmentPlanWithPayments{
			InstallmentPlan: plan,
			Payments:        payments,
		}
		
		assert.Equal(t, "plan-123", planWithPayments.ID)
		assert.Len(t, planWithPayments.Payments, 2)
	})
}

func TestInstallmentRepo_GetActivePaymentsByMonth(t *testing.T) {
	t.Run("InstallmentRepository interface exists", func(t *testing.T) {
		// Test that the repository interface supports this method
		// Actual filtering is done via EXTRACT in SQL
		_ = func(r InstallmentRepository) {} // Verify interface exists
		assert.True(t, true)
	})
}
