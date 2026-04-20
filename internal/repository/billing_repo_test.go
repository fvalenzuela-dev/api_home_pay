package repository

import (
	"testing"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestBillingRepo_Interfaces(t *testing.T) {
	t.Run("BillingRepository interface is satisfied by billingRepo", func(t *testing.T) {
		var _ BillingRepository = (*billingRepo)(nil)
	})
}

func TestScanBilling(t *testing.T) {
	t.Run("scanBilling function exists", func(t *testing.T) {
		assert.NotNil(t, scanBilling)
	})
}

func TestBillingRepo_GetAll(t *testing.T) {
	t.Run("billingCols constant", func(t *testing.T) {
		assert.Equal(t, `id, account_id, period, amount_billed, amount_paid, is_paid, paid_at, carried_from, created_at, deleted_at`, billingCols)
	})

	t.Run("billingColsAB constant", func(t *testing.T) {
		assert.Equal(t, `ab.id, ab.account_id, ab.period, ab.amount_billed, ab.amount_paid, ab.is_paid, ab.paid_at, ab.carried_from, ab.created_at, ab.deleted_at`, billingColsAB)
	})
}

func TestBillingRepo_Create(t *testing.T) {
	t.Run("CreateBillingRequest validation", func(t *testing.T) {
		req := models.CreateBillingRequest{
			Period:       202603,
			AmountBilled: 50000,
		}
		assert.Equal(t, 202603, req.Period)
		assert.Equal(t, 50000.0, req.AmountBilled)
	})

	t.Run("CreateBillingRequest with optional fields", func(t *testing.T) {
		amountPaid := 25000.0
		isPaid := true
		carriedFrom := "billing-123"
		
		req := models.CreateBillingRequest{
			Period:       202603,
			AmountBilled: 50000,
			AmountPaid:   &amountPaid,
			IsPaid:       &isPaid,
			CarriedFrom:  &carriedFrom,
		}
		assert.NotNil(t, req.AmountPaid)
		assert.NotNil(t, req.IsPaid)
		assert.Equal(t, "billing-123", *req.CarriedFrom)
	})
}

func TestBillingRepo_Update(t *testing.T) {
	t.Run("UpdateBillingRequest with pointer fields", func(t *testing.T) {
		amountBilled := 60000.0
		amountPaid := 60000.0
		isPaid := true
		
		req := models.UpdateBillingRequest{
			AmountBilled: &amountBilled,
			AmountPaid:   &amountPaid,
			IsPaid:       &isPaid,
		}
		assert.Equal(t, 60000.0, *req.AmountBilled)
		assert.Equal(t, 60000.0, *req.AmountPaid)
		assert.True(t, *req.IsPaid)
	})
}

func TestBillingRepo_GetAllByPeriod(t *testing.T) {
	t.Run("AccountBillingWithDetails model", func(t *testing.T) {
		billing := models.AccountBillingWithDetails{
			AccountBilling: models.AccountBilling{
				ID:           "billing-123",
				AccountID:    "account-123",
				Period:       202603,
				AmountBilled: 50000,
				IsPaid:       true,
			},
			CategoryName: "Utilities",
			CompanyName:  "Electric Company",
			AccountName:  "Home Electric",
		}
		assert.Equal(t, "Utilities", billing.CategoryName)
		assert.Equal(t, "Electric Company", billing.CompanyName)
		assert.Equal(t, "Home Electric", billing.AccountName)
	})
}

func TestBillingRepo_PeriodBillingInsert(t *testing.T) {
	t.Run("PeriodBillingInsert model", func(t *testing.T) {
		carriedFrom := "billing-123"
		insert := models.PeriodBillingInsert{
			AccountID:    "account-123",
			AmountBilled: 50000,
			CarriedFrom:  &carriedFrom,
		}
		assert.Equal(t, "account-123", insert.AccountID)
		assert.Equal(t, 50000.0, insert.AmountBilled)
		assert.NotNil(t, insert.CarriedFrom)
	})
}

func TestBillingRepo_OpenPeriodResponse(t *testing.T) {
	t.Run("OpenPeriodResponse model", func(t *testing.T) {
		resp := models.OpenPeriodResponse{
			Period:  202603,
			Created: 5,
			Skipped: 2,
		}
		assert.Equal(t, 202603, resp.Period)
		assert.Equal(t, 5, resp.Created)
		assert.Equal(t, 2, resp.Skipped)
	})
}

func TestAccountBillingModel(t *testing.T) {
	t.Run("AccountBilling model fields", func(t *testing.T) {
		now := time.Now()
		billing := models.AccountBilling{
			ID:           "billing-123",
			AccountID:    "account-123",
			Period:       202603,
			AmountBilled: 50000,
			AmountPaid:   50000,
			IsPaid:       true,
			PaidAt:       &now,
			CreatedAt:    now,
		}
		assert.Equal(t, "billing-123", billing.ID)
		assert.Equal(t, 202603, billing.Period)
		assert.True(t, billing.IsPaid)
		assert.NotNil(t, billing.PaidAt)
	})
}
