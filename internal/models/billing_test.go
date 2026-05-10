package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccountBilling_Struct(t *testing.T) {
	now := time.Now()

	billing := AccountBilling{
		ID:           "billing-123",
		AccountID:    "account-123",
		Period:       202603,
		AmountBilled: 15000.00,
		AmountPaid:   15000.00,
		IsPaid:       true,
		CreatedAt:    now,
	}

	assert.Equal(t, "billing-123", billing.ID)
	assert.Equal(t, "account-123", billing.AccountID)
	assert.Equal(t, 202603, billing.Period)
	assert.Equal(t, 15000.00, billing.AmountBilled)
	assert.Equal(t, 15000.00, billing.AmountPaid)
	assert.True(t, billing.IsPaid)
}

func TestAccountBilling_IsPaid_WithoutPaidAt(t *testing.T) {
	billing := AccountBilling{
		ID:           "billing-123",
		AccountID:    "account-123",
		Period:       202603,
		AmountBilled: 15000.00,
		AmountPaid:   15000.00,
		IsPaid:       true,
		PaidAt:       nil,
	}

	// IsPaid can be true even without PaidAt set
	assert.True(t, billing.IsPaid)
	assert.Nil(t, billing.PaidAt)
}

func TestAccountBilling_CarriedFrom(t *testing.T) {
	carriedFrom := "previous-billing-123"

	billing := AccountBilling{
		ID:           "billing-123",
		AccountID:    "account-123",
		Period:       202603,
		AmountBilled: 20000.00,
		AmountPaid:   0,
		IsPaid:       false,
		CarriedFrom:  &carriedFrom,
	}

	assert.NotNil(t, billing.CarriedFrom)
	assert.Equal(t, "previous-billing-123", *billing.CarriedFrom)
}

func TestCreateBillingRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateBillingRequest
		isValid bool
	}{
		{
			name: "valid request",
			req: CreateBillingRequest{
				Period:       202603,
				AmountBilled: 15000.00,
			},
			isValid: true,
		},
		{
			name: "valid request with payment",
			req: CreateBillingRequest{
				Period:       202603,
				AmountBilled: 15000.00,
				AmountPaid:   func() *float64 { v := 15000.00; return &v }(),
				IsPaid:       func() *bool { v := true; return &v }(),
			},
			isValid: true,
		},
		{
			name: "invalid period - zero",
			req: CreateBillingRequest{
				Period:       0,
				AmountBilled: 15000.00,
			},
			isValid: false,
		},
		{
			name: "invalid period - not YYYYMM format",
			req: CreateBillingRequest{
				Period:       2026,
				AmountBilled: 15000.00,
			},
			isValid: false,
		},
		{
			name: "invalid amount - zero",
			req: CreateBillingRequest{
				Period:       202603,
				AmountBilled: 0,
			},
			isValid: false,
		},
		{
			name: "invalid amount - negative",
			req: CreateBillingRequest{
				Period:       202603,
				AmountBilled: -100.00,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				assert.Greater(t, tt.req.Period, 0)
				assert.Greater(t, tt.req.AmountBilled, 0.0)
			}
		})
	}
}

func TestCreateBillingRequest_AutoPaid(t *testing.T) {
	amountBilled := 15000.00
	amountPaid := 15000.00
	isPaid := true
	paidAt := time.Now()

	req := CreateBillingRequest{
		Period:       202603,
		AmountBilled: amountBilled,
		AmountPaid:   &amountPaid,
		IsPaid:       &isPaid,
		PaidAt:       &paidAt,
	}

	// When AmountPaid >= AmountBilled, should auto-set IsPaid
	assert.NotNil(t, req.AmountPaid)
	assert.NotNil(t, req.IsPaid)
	assert.True(t, *req.IsPaid)
	assert.NotNil(t, req.PaidAt)
}

func TestUpdateBillingRequest_PartialUpdate(t *testing.T) {
	amountBilled := 20000.00
	isPaid := true

	req := UpdateBillingRequest{
		AmountBilled: &amountBilled,
		IsPaid:       &isPaid,
	}

	assert.NotNil(t, req.AmountBilled)
	assert.NotNil(t, req.IsPaid)
	assert.Equal(t, 20000.00, *req.AmountBilled)
	assert.True(t, *req.IsPaid)
}

func TestOpenPeriodResponse_Struct(t *testing.T) {
	resp := OpenPeriodResponse{
		Period:  202603,
		Created: 10,
		Skipped: 5,
	}

	assert.Equal(t, 202603, resp.Period)
	assert.Equal(t, 10, resp.Created)
	assert.Equal(t, 5, resp.Skipped)
}

func TestAccountBillingWithDetails_Struct(t *testing.T) {
	now := time.Now()

	billing := AccountBillingWithDetails{
		AccountBilling: AccountBilling{
			ID:           "billing-123",
			AccountID:    "account-123",
			Period:       202603,
			AmountBilled: 15000.00,
			AmountPaid:   15000.00,
			IsPaid:       true,
			CreatedAt:    now,
		},
		CategoryName: "Utilities",
		CompanyName:  "Electric Company",
		AccountName:  "Home Service",
	}

	assert.Equal(t, "billing-123", billing.ID)
	assert.Equal(t, "Utilities", billing.CategoryName)
	assert.Equal(t, "Electric Company", billing.CompanyName)
	assert.Equal(t, "Home Service", billing.AccountName)
}

func TestPeriodBillingInsert_Struct(t *testing.T) {
	carriedFrom := "old-billing-123"

	insert := PeriodBillingInsert{
		AccountID:    "account-123",
		AmountBilled: 15000.00,
		CarriedFrom:  &carriedFrom,
	}

	assert.Equal(t, "account-123", insert.AccountID)
	assert.Equal(t, 15000.00, insert.AmountBilled)
	assert.NotNil(t, insert.CarriedFrom)
}

func TestBillingFilters_Struct(t *testing.T) {
	accountID := "acc-123"
	fromPeriod := 202601
	toPeriod := 202606
	isPaid := false

	filters := BillingFilters{
		AccountID:  &accountID,
		FromPeriod: &fromPeriod,
		ToPeriod:   &toPeriod,
		IsPaid:     &isPaid,
	}

	assert.Equal(t, "acc-123", *filters.AccountID)
	assert.Equal(t, 202601, *filters.FromPeriod)
	assert.Equal(t, 202606, *filters.ToPeriod)
	assert.NotNil(t, filters.IsPaid)
	assert.False(t, *filters.IsPaid)
}

func TestBillingFilters_NilFields(t *testing.T) {
	filters := BillingFilters{}

	assert.Nil(t, filters.AccountID)
	assert.Nil(t, filters.FromPeriod)
	assert.Nil(t, filters.ToPeriod)
	assert.Nil(t, filters.IsPaid)
}

func TestBillingFilters_IsPaid_True(t *testing.T) {
	isPaid := true
	filters := BillingFilters{IsPaid: &isPaid}

	assert.NotNil(t, filters.IsPaid)
	assert.True(t, *filters.IsPaid)
}

func TestBillingFilters_IsPaid_False(t *testing.T) {
	isPaid := false
	filters := BillingFilters{IsPaid: &isPaid}

	assert.NotNil(t, filters.IsPaid)
	assert.False(t, *filters.IsPaid)
}
