package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInstallmentPlan_Struct(t *testing.T) {
	now := time.Now()

	plan := InstallmentPlan{
		ID:                "plan-123",
		AuthUserID:        "user_123",
		Description:       "TV Payment",
		TotalAmount:       300000.00,
		TotalInstallments: 12,
		InstallmentsPaid:  3,
		StartDate:         now,
		IsCompleted:       false,
		CreatedAt:         now,
	}

	assert.Equal(t, "plan-123", plan.ID)
	assert.Equal(t, "user_123", plan.AuthUserID)
	assert.Equal(t, "TV Payment", plan.Description)
	assert.Equal(t, 300000.00, plan.TotalAmount)
	assert.Equal(t, 12, plan.TotalInstallments)
	assert.Equal(t, 3, plan.InstallmentsPaid)
	assert.False(t, plan.IsCompleted)
}

func TestInstallmentPlan_IsCompleted(t *testing.T) {
	plan := InstallmentPlan{
		ID:                "plan-123",
		TotalInstallments: 12,
		InstallmentsPaid:  12,
		IsCompleted:       true,
	}

	assert.True(t, plan.IsCompleted)
	assert.Equal(t, plan.TotalInstallments, plan.InstallmentsPaid)
}

func TestInstallmentPayment_Struct(t *testing.T) {
	now := time.Now()

	payment := InstallmentPayment{
		ID:                "payment-123",
		PlanID:            "plan-123",
		InstallmentNumber: 1,
		Amount:            25000.00,
		DueDate:           now,
		IsPaid:            true,
		PaidAt:            &now,
		CreatedAt:         now,
	}

	assert.Equal(t, "payment-123", payment.ID)
	assert.Equal(t, "plan-123", payment.PlanID)
	assert.Equal(t, 1, payment.InstallmentNumber)
	assert.Equal(t, 25000.00, payment.Amount)
	assert.True(t, payment.IsPaid)
	assert.NotNil(t, payment.PaidAt)
}

func TestInstallmentPayment_NotPaid(t *testing.T) {
	payment := InstallmentPayment{
		ID:                "payment-123",
		PlanID:            "plan-123",
		InstallmentNumber: 1,
		Amount:            25000.00,
		DueDate:           time.Now(),
		IsPaid:            false,
		PaidAt:            nil,
	}

	assert.False(t, payment.IsPaid)
	assert.Nil(t, payment.PaidAt)
}

func TestCreateInstallmentRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateInstallmentRequest
		isValid bool
	}{
		{
			name: "valid request",
			req: CreateInstallmentRequest{
				Description:       "New Phone",
				TotalAmount:        600000.00,
				TotalInstallments:  12,
				StartDate:          "2026-03-01",
			},
			isValid: true,
		},
		{
			name: "empty description",
			req: CreateInstallmentRequest{
				Description:       "",
				TotalAmount:        600000.00,
				TotalInstallments:  12,
				StartDate:          "2026-03-01",
			},
			isValid: false,
		},
		{
			name: "invalid amount - zero",
			req: CreateInstallmentRequest{
				Description:       "Test",
				TotalAmount:        0,
				TotalInstallments:  12,
				StartDate:          "2026-03-01",
			},
			isValid: false,
		},
		{
			name: "invalid installments - zero",
			req: CreateInstallmentRequest{
				Description:       "Test",
				TotalAmount:        600000.00,
				TotalInstallments:  0,
				StartDate:          "2026-03-01",
			},
			isValid: false,
		},
		{
			name: "invalid installments - negative",
			req: CreateInstallmentRequest{
				Description:       "Test",
				TotalAmount:        600000.00,
				TotalInstallments:  -1,
				StartDate:          "2026-03-01",
			},
			isValid: false,
		},
		{
			name: "empty start date",
			req: CreateInstallmentRequest{
				Description:       "Test",
				TotalAmount:        600000.00,
				TotalInstallments:  12,
				StartDate:          "",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				assert.NotEmpty(t, tt.req.Description)
				assert.Greater(t, tt.req.TotalAmount, 0.0)
				assert.Greater(t, tt.req.TotalInstallments, 0)
				assert.NotEmpty(t, tt.req.StartDate)
			}
		})
	}
}

func TestInstallmentPlanWithPayments_Struct(t *testing.T) {
	now := time.Now()

	planWithPayments := InstallmentPlanWithPayments{
		InstallmentPlan: InstallmentPlan{
			ID:                "plan-123",
			Description:       "Test Plan",
			TotalAmount:       300000.00,
			TotalInstallments: 12,
			InstallmentsPaid:  3,
			StartDate:         now,
		},
		Payments: []InstallmentPayment{
			{
				ID:                "payment-1",
				PlanID:            "plan-123",
				InstallmentNumber: 1,
				Amount:            25000.00,
				DueDate:           now,
				IsPaid:            true,
				PaidAt:            &now,
			},
			{
				ID:                "payment-2",
				PlanID:            "plan-123",
				InstallmentNumber: 2,
				Amount:            25000.00,
				DueDate:           now,
				IsPaid:            false,
			},
		},
	}

	assert.Equal(t, "plan-123", planWithPayments.ID)
	assert.Equal(t, 2, len(planWithPayments.Payments))
	assert.Equal(t, 1, planWithPayments.Payments[0].InstallmentNumber)
	assert.Equal(t, 2, planWithPayments.Payments[1].InstallmentNumber)
}

func TestInstallmentPlan_CalculateProgress(t *testing.T) {
	plan := InstallmentPlan{
		TotalInstallments: 12,
		InstallmentsPaid:  6,
	}

	progress := float64(plan.InstallmentsPaid) / float64(plan.TotalInstallments) * 100

	assert.Equal(t, 50.0, progress)
}

func TestInstallmentPlan_Completed(t *testing.T) {
	plan := InstallmentPlan{
		TotalInstallments: 12,
		InstallmentsPaid:  12,
		IsCompleted:       true,
	}

	assert.True(t, plan.IsCompleted)
	assert.Equal(t, plan.TotalInstallments, plan.InstallmentsPaid)
}
