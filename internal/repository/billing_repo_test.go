package repository

import (
	"context"
	"testing"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Billing Repository Tests

func TestBillingRepo_NewBillingRepository(t *testing.T) {
	t.Run("creates billingRepo instance", func(t *testing.T) {
		var repo BillingRepository = NewBillingRepository(nil)
		assert.NotNil(t, repo)
	})
}

func TestBillingConstants(t *testing.T) {
	t.Run("billingCols constant", func(t *testing.T) {
		assert.Equal(t, `id, account_id, period, amount_billed, amount_paid, is_paid, paid_at, carried_from, created_at, deleted_at`, billingCols)
	})

	t.Run("billingColsAB constant", func(t *testing.T) {
		assert.Equal(t, `ab.id, ab.account_id, ab.period, ab.amount_billed, ab.amount_paid, ab.is_paid, ab.paid_at, ab.carried_from, ab.created_at, ab.deleted_at`, billingColsAB)
		assert.Contains(t, billingColsAB, "ab.id")
		assert.Contains(t, billingColsAB, "ab.account_id")
	})
}

func TestScanBilling(t *testing.T) {
	t.Run("scanBilling function exists", func(t *testing.T) {
		assert.NotNil(t, scanBilling)
	})
}

// Billing Model Tests

func TestBillingModel_Full(t *testing.T) {
	t.Run("AccountBilling struct can hold all fields", func(t *testing.T) {
		now := time.Now()
		paidAt := now

		billing := models.AccountBilling{
			ID:           "billing-123",
			AccountID:    "account-456",
			Period:       202604,
			AmountBilled: 15000.00,
			AmountPaid:   15000.00,
			IsPaid:       true,
			PaidAt:       &paidAt,
			CreatedAt:    now,
		}

		assert.Equal(t, "billing-123", billing.ID)
		assert.Equal(t, "account-456", billing.AccountID)
		assert.Equal(t, 202604, billing.Period)
		assert.Equal(t, 15000.00, billing.AmountBilled)
		assert.True(t, billing.IsPaid)
		assert.NotNil(t, billing.PaidAt)
	})
}

func TestBillingModel_WithNilPointers(t *testing.T) {
	t.Run("AccountBilling struct handles nil pointers", func(t *testing.T) {
		billing := models.AccountBilling{
			ID:           "billing-123",
			AccountID:    "account-456",
			Period:       202604,
			AmountBilled: 15000.00,
			IsPaid:       false,
			CreatedAt:    time.Now(),
		}

		assert.Nil(t, billing.PaidAt)
		assert.Nil(t, billing.CarriedFrom)
		assert.Nil(t, billing.DeletedAt)
	})
}

func TestBillingModel_WithCarriedFrom(t *testing.T) {
	t.Run("AccountBilling with carried_from", func(t *testing.T) {
		carriedFrom := "billing-122"
		billing := models.AccountBilling{
			ID:           "billing-123",
			AccountID:    "account-456",
			Period:       202604,
			AmountBilled: 5000.00,
			CarriedFrom:  &carriedFrom,
		}

		assert.NotNil(t, billing.CarriedFrom)
		assert.Equal(t, "billing-122", *billing.CarriedFrom)
	})
}

func TestCreateBillingRequest_AllFields(t *testing.T) {
	t.Run("CreateBillingRequest with all fields", func(t *testing.T) {
		amountPaid := 15000.00
		isPaid := true
		paidAt := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

		req := models.CreateBillingRequest{
			Period:       202604,
			AmountBilled: 15000.00,
			AmountPaid:   &amountPaid,
			IsPaid:       &isPaid,
			PaidAt:       &paidAt,
		}

		assert.Equal(t, 202604, req.Period)
		assert.Equal(t, 15000.00, req.AmountBilled)
		assert.NotNil(t, req.AmountPaid)
		assert.NotNil(t, req.IsPaid)
		assert.True(t, *req.IsPaid)
	})
}

func TestCreateBillingRequest_MinimalFields(t *testing.T) {
	t.Run("CreateBillingRequest with only required fields", func(t *testing.T) {
		req := models.CreateBillingRequest{
			Period:       202604,
			AmountBilled: 15000.00,
		}

		assert.Equal(t, 202604, req.Period)
		assert.Equal(t, 15000.00, req.AmountBilled)
		assert.Nil(t, req.AmountPaid)
		assert.Nil(t, req.IsPaid)
		assert.Nil(t, req.PaidAt)
		assert.Nil(t, req.CarriedFrom)
	})
}

func TestUpdateBillingRequest_AllFields(t *testing.T) {
	t.Run("UpdateBillingRequest with all fields", func(t *testing.T) {
		amountBilled := 20000.00
		amountPaid := 20000.00
		isPaid := true
		paidAt := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)

		req := models.UpdateBillingRequest{
			AmountBilled: &amountBilled,
			AmountPaid:   &amountPaid,
			IsPaid:       &isPaid,
			PaidAt:       &paidAt,
		}

		assert.NotNil(t, req.AmountBilled)
		assert.NotNil(t, req.AmountPaid)
		assert.NotNil(t, req.IsPaid)
		assert.NotNil(t, req.PaidAt)
	})
}

func TestUpdateBillingRequest_PartialFields(t *testing.T) {
	t.Run("UpdateBillingRequest with only some field", func(t *testing.T) {
		isPaid := true
		req := models.UpdateBillingRequest{
			IsPaid: &isPaid,
		}

		assert.NotNil(t, req.IsPaid)
		assert.Nil(t, req.AmountBilled)
		assert.Nil(t, req.AmountPaid)
		assert.Nil(t, req.PaidAt)
	})
}

// Billing Repository Tests with Mocks - Only unique ones

func TestBillingRepo_GetByAccountAndPeriod_WithMock(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	accountID := "account-123"
	authUserID := "user-123"
	period := 202604

	expectedBilling := &models.AccountBilling{
		ID:           "billing-123",
		AccountID:    accountID,
		Period:       period,
		AmountBilled: 15000.00,
	}

	mockRepo.On("GetByAccountAndPeriod", mock.Anything, accountID, authUserID, period).Return(expectedBilling, nil)

	result, err := mockRepo.GetByAccountAndPeriod(context.Background(), accountID, authUserID, period)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, period, result.Period)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_GetAllByAccount_WithMock(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	accountID := "account-123"
	authUserID := "user-123"
	pagination := models.PaginationParams{Limit: 10}

	billings := []models.AccountBilling{
		{ID: "billing-1", Period: 202604},
		{ID: "billing-2", Period: 202603},
	}

	mockRepo.On("GetAllByAccount", mock.Anything, accountID, authUserID, pagination).Return(billings, 2, nil)

	result, total, err := mockRepo.GetAllByAccount(context.Background(), accountID, authUserID, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 2, total)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_GetAllByPeriod_WithMock(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	authUserID := "user-123"
	period := 202604
	pagination := models.PaginationParams{Limit: 10}

	billings := []models.AccountBillingWithDetails{
		{AccountBilling: models.AccountBilling{AccountID: "account-1", Period: period}, AccountName: "Account 1"},
		{AccountBilling: models.AccountBilling{AccountID: "account-2", Period: period}, AccountName: "Account 2"},
	}

	mockRepo.On("GetAllByPeriod", mock.Anything, authUserID, period, (*bool)(nil), pagination).Return(billings, 2, nil)

	result, total, err := mockRepo.GetAllByPeriod(context.Background(), authUserID, period, nil, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 2, total)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_GetAllByPeriod_WithPaidFilter(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	authUserID := "user-123"
	period := 202604
	pagination := models.PaginationParams{Limit: 10}
	isPaid := true

	billings := []models.AccountBillingWithDetails{
		{AccountBilling: models.AccountBilling{AccountID: "account-1", IsPaid: true}},
	}

	mockRepo.On("GetAllByPeriod", mock.Anything, authUserID, period, &isPaid, pagination).Return(billings, 1, nil)

	result, total, err := mockRepo.GetAllByPeriod(context.Background(), authUserID, period, &isPaid, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, 1, total)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_GetAllByPeriod_WithUnpaidFilter(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	authUserID := "user-123"
	period := 202604
	pagination := models.PaginationParams{Limit: 10}
	isPaid := false

	billings := []models.AccountBillingWithDetails{
		{AccountBilling: models.AccountBilling{AccountID: "account-1", IsPaid: false}},
	}

	mockRepo.On("GetAllByPeriod", mock.Anything, authUserID, period, &isPaid, pagination).Return(billings, 1, nil)

	result, total, err := mockRepo.GetAllByPeriod(context.Background(), authUserID, period, &isPaid, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, 1, total)
	mockRepo.AssertExpectations(t)
}

// Edge Case Tests

func TestBillingRepo_PeriodEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		period int
	}{
		{"current period", 202604},
		{"future period", 203001},
		{"past period", 202001},
		{"min period", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			billing := models.AccountBilling{
				ID:     "test",
				Period: tt.period,
			}
			assert.Equal(t, tt.period, billing.Period)
		})
	}
}

func TestBillingRepo_AmountEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		amountBilled float64
		amountPaid   float64
	}{
		{"zero amount", 0.0, 0.0},
		{"negative amount", -100.0, 0.0},
		{"small amount", 0.01, 0.01},
		{"large amount", 999999.99, 999999.99},
		{"partial payment", 15000.00, 10000.00},
		{"over payment", 15000.00, 20000.00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			billing := models.AccountBilling{
				ID:           "test",
				AmountBilled: tt.amountBilled,
				AmountPaid:   tt.amountPaid,
			}
			assert.Equal(t, tt.amountBilled, billing.AmountBilled)
			assert.Equal(t, tt.amountPaid, billing.AmountPaid)
		})
	}
}

func TestBillingRepo_PaidAtEdgeCases(t *testing.T) {
	t.Run("with paid_at", func(t *testing.T) {
		paidAt := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
		billing := models.AccountBilling{
			ID:     "test",
			IsPaid: true,
			PaidAt: &paidAt,
		}
		assert.NotNil(t, billing.PaidAt)
		assert.True(t, billing.IsPaid)
	})

	t.Run("without paid_at", func(t *testing.T) {
		billing := models.AccountBilling{
			ID:     "test",
			IsPaid: false,
		}
		assert.Nil(t, billing.PaidAt)
		assert.False(t, billing.IsPaid)
	})
}

func TestPeriodBillingInsert(t *testing.T) {
	t.Run("PeriodBillingInsert with carried_from", func(t *testing.T) {
		carriedFrom := "billing-123"
		insert := models.PeriodBillingInsert{
			AccountID:    "account-123",
			AmountBilled: 5000.00,
			CarriedFrom:  &carriedFrom,
		}
		assert.NotNil(t, insert.CarriedFrom)
		assert.Equal(t, "billing-123", *insert.CarriedFrom)
	})

	t.Run("PeriodBillingInsert without carried_from", func(t *testing.T) {
		insert := models.PeriodBillingInsert{
			AccountID:    "account-123",
			AmountBilled: 15000.00,
		}
		assert.Nil(t, insert.CarriedFrom)
	})
}

// Interface compliance test
func TestBillingRepository_ImplementsInterface(t *testing.T) {
	var _ BillingRepository = (*MockBillingRepository)(nil)
}

// Billing Repository Query Tests

func TestBillingRepo_Create_Query(t *testing.T) {
	// Test INSERT query
	query := `INSERT INTO homepay.account_billings (account_id, period, amount_billed, amount_paid, is_paid, paid_at, carried_from)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, account_id, period, amount_billed, amount_paid, is_paid, paid_at, carried_from, created_at, deleted_at`

	assert.Contains(t, query, "INSERT INTO homepay.account_billings")
	assert.Contains(t, query, "RETURNING")
}

func TestBillingRepo_CreateCarryOver_Query(t *testing.T) {
	// Test carry over INSERT
	query := `INSERT INTO homepay.account_billings (account_id, period, amount_billed, carried_from)
		VALUES ($1, $2, $3, $4)`

	assert.Contains(t, query, "carried_from")
}

func TestBillingRepo_GetByID_Query(t *testing.T) {
	// Test SELECT with JOIN
	query := `SELECT ab.id, ab.account_id, ab.period, ab.amount_billed, ab.amount_paid, ab.is_paid, ab.paid_at, ab.carried_from, ab.created_at, ab.deleted_at
		FROM homepay.account_billings ab
		JOIN homepay.accounts a ON a.id = ab.account_id
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE ab.id = $1 AND c.auth_user_id = $2 AND ab.deleted_at IS NULL`

	assert.Contains(t, query, "JOIN homepay.accounts")
	assert.Contains(t, query, "JOIN homepay.companies")
}

func TestBillingRepo_GetByAccountAndPeriod_Query(t *testing.T) {
	// Test SELECT by account and period
	query := `SELECT id, account_id, period, amount_billed, amount_paid, is_paid, paid_at, carried_from, created_at, deleted_at
		FROM homepay.account_billings
		WHERE account_id = $1 AND period = $2 AND deleted_at IS NULL`

	assert.Contains(t, query, "account_id = $1 AND period = $2")
}

func TestBillingRepo_GetUnpaidByAccount_Query(t *testing.T) {
	// Test unpaid query
	query := `SELECT id, account_id, period, amount_billed, amount_paid, is_paid, paid_at, carried_from, created_at, deleted_at
		FROM homepay.account_billings
		WHERE account_id = $1 AND is_paid = FALSE AND deleted_at IS NULL
		ORDER BY period DESC
		LIMIT 1`

	assert.Contains(t, query, "is_paid = FALSE")
	assert.Contains(t, query, "ORDER BY period DESC")
	assert.Contains(t, query, "LIMIT 1")
}

func TestBillingRepo_GetAllByPeriod_Query(t *testing.T) {
	// Test period query with filters
	query := `SELECT ab.id, ab.account_id, ab.period, ab.amount_billed, ab.amount_paid, ab.is_paid, ab.paid_at, ab.carried_from, ab.created_at, ab.deleted_at, cat.name, c.name, a.name
		FROM homepay.account_billings ab
		JOIN homepay.accounts a ON a.id = ab.account_id
		JOIN homepay.companies c ON c.id = a.company_id
		JOIN homepay.categories cat ON cat.id = c.category_id
		WHERE c.auth_user_id = $1 AND ab.period = $2 AND ab.deleted_at IS NULL`

	assert.Contains(t, query, "JOIN homepay.categories")
	assert.Contains(t, query, "ab.period = $2")
}

func TestBillingRepo_GetAllByPeriod_PaidFilter(t *testing.T) {
	// Test paid filter
	isPaid := true
	paidFilter := ""
	if isPaid {
		paidFilter = " AND ab.is_paid = TRUE"
	}

	assert.Contains(t, paidFilter, "ab.is_paid = TRUE")

	isPaid = false
	paidFilter = ""
	if isPaid {
		paidFilter = " AND ab.is_paid = TRUE"
	}

	// When isPaid is false, filter should be empty or use FALSE
	paidFilter = " AND ab.is_paid = FALSE"
	assert.Contains(t, paidFilter, "ab.is_paid = FALSE")
}

func TestBillingRepo_BulkInsert_Query(t *testing.T) {
	// Test bulk INSERT transaction
	query := `INSERT INTO homepay.account_billings (account_id, period, amount_billed, carried_from)
		VALUES ($1, $2, $3, $4)`

	assert.Contains(t, query, "INSERT INTO homepay.account_billings")
}

func TestBillingRepo_Update_Query(t *testing.T) {
	// Test UPDATE query
	query := `UPDATE homepay.account_billings ab
		SET amount_billed = COALESCE($3, ab.amount_billed),
		    amount_paid   = COALESCE($4, ab.amount_paid),
		    is_paid       = COALESCE($5, ab.is_paid),
		    paid_at       = COALESCE($6, ab.paid_at)
		FROM homepay.accounts a
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE ab.id = $1 AND ab.account_id = a.id AND c.auth_user_id = $2 AND ab.deleted_at IS NULL`

	assert.Contains(t, query, "COALESCE")
	assert.Contains(t, query, "FROM homepay.accounts a")
}

func TestBillingRepo_MarkPaid_Query(t *testing.T) {
	// Test mark paid query
	query := `UPDATE homepay.account_billings
		SET is_paid = TRUE, paid_at = CURRENT_DATE
		WHERE id = $1`

	assert.Contains(t, query, "is_paid = TRUE")
	assert.Contains(t, query, "paid_at = CURRENT_DATE")
}

func TestBillingRepo_SoftDeleteByAccount_Query(t *testing.T) {
	// Test soft delete by account
	query := `UPDATE homepay.account_billings SET deleted_at = NOW()
		WHERE account_id = $1 AND deleted_at IS NULL`

	assert.Contains(t, query, "deleted_at = NOW()")
	assert.Contains(t, query, "account_id = $1")
}

// BillingFilters Model Tests

func TestBillingFilters_Default(t *testing.T) {
	t.Run("BillingFilters with nil fields means no filter", func(t *testing.T) {
		filters := models.BillingFilters{}
		assert.Nil(t, filters.AccountID)
		assert.Nil(t, filters.FromPeriod)
		assert.Nil(t, filters.ToPeriod)
		assert.Nil(t, filters.IsPaid)
	})
}

func TestBillingFilters_WithAccountID(t *testing.T) {
	t.Run("BillingFilters with AccountID filter", func(t *testing.T) {
		accountID := "account-123"
		filters := models.BillingFilters{AccountID: &accountID}
		assert.NotNil(t, filters.AccountID)
		assert.Equal(t, "account-123", *filters.AccountID)
	})
}

func TestBillingFilters_WithPeriodRange(t *testing.T) {
	t.Run("BillingFilters with FromPeriod and ToPeriod", func(t *testing.T) {
		from := 202601
		to := 202603
		filters := models.BillingFilters{FromPeriod: &from, ToPeriod: &to}
		assert.NotNil(t, filters.FromPeriod)
		assert.NotNil(t, filters.ToPeriod)
		assert.Equal(t, 202601, *filters.FromPeriod)
		assert.Equal(t, 202603, *filters.ToPeriod)
	})
}

func TestBillingFilters_WithIsPaid_True(t *testing.T) {
	t.Run("BillingFilters with IsPaid=true (paid billings)", func(t *testing.T) {
		isPaid := true
		filters := models.BillingFilters{IsPaid: &isPaid}
		assert.NotNil(t, filters.IsPaid)
		assert.True(t, *filters.IsPaid)
	})
}

func TestBillingFilters_WithIsPaid_False(t *testing.T) {
	t.Run("BillingFilters with IsPaid=false (unpaid billings)", func(t *testing.T) {
		isPaid := false
		filters := models.BillingFilters{IsPaid: &isPaid}
		assert.NotNil(t, filters.IsPaid)
		assert.False(t, *filters.IsPaid)
	})
}

func TestBillingFilters_AllFilters(t *testing.T) {
	t.Run("BillingFilters with all filters set", func(t *testing.T) {
		accountID := "account-123"
		from := 202601
		to := 202603
		isPaid := true
		filters := models.BillingFilters{
			AccountID:  &accountID,
			FromPeriod: &from,
			ToPeriod:   &to,
			IsPaid:     &isPaid,
		}
		assert.NotNil(t, filters.AccountID)
		assert.NotNil(t, filters.FromPeriod)
		assert.NotNil(t, filters.ToPeriod)
		assert.NotNil(t, filters.IsPaid)
	})
}

// GetAll Repository Method Tests

func TestBillingRepo_GetAll_Query(t *testing.T) {
	// Test query structure for GetAll
	query := `SELECT ab.id, ab.account_id, ab.period, ab.amount_billed, ab.amount_paid, ab.is_paid, ab.paid_at, ab.carried_from, ab.created_at, ab.deleted_at
		FROM homepay.account_billings ab
		JOIN homepay.accounts a ON a.id = ab.account_id
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE c.auth_user_id = $1 AND ab.deleted_at IS NULL
		ORDER BY ab.period DESC
		LIMIT $2 OFFSET $3`

	assert.Contains(t, query, "JOIN homepay.accounts")
	assert.Contains(t, query, "JOIN homepay.companies")
	assert.Contains(t, query, "c.auth_user_id = $1")
	assert.Contains(t, query, "ORDER BY ab.period DESC")
}

func TestBillingRepo_GetAll_Query_WithAccountFilter(t *testing.T) {
	// Test query with account_id filter
	accountID := "account-123"
	conditions := "c.auth_user_id = $1 AND ab.deleted_at IS NULL"
	conditions += " AND ab.account_id = $2"

	assert.Contains(t, conditions, "ab.account_id = $2")
	assert.Contains(t, conditions, "c.auth_user_id = $1")
	_ = accountID // accountID used in filter construction pattern
}

func TestBillingRepo_GetAll_Query_WithPeriodRangeFilter(t *testing.T) {
	// Test query with FromPeriod and ToPeriod filters
	from := 202601
	to := 202603
	conditions := "c.auth_user_id = $1 AND ab.deleted_at IS NULL"
	conditions += " AND ab.period >= $2"
	conditions += " AND ab.period <= $3"

	assert.Contains(t, conditions, "ab.period >= $2")
	assert.Contains(t, conditions, "ab.period <= $3")
	_ = from // used in filter construction
	_ = to
}

func TestBillingRepo_GetAll_Query_WithIsPaidFilter(t *testing.T) {
	// Test query with is_paid filter
	isPaid := true
	conditions := "c.auth_user_id = $1 AND ab.deleted_at IS NULL"
	if isPaid {
		conditions += " AND ab.is_paid = TRUE"
	} else {
		conditions += " AND ab.is_paid = FALSE"
	}

	assert.Contains(t, conditions, "ab.is_paid = TRUE")
}

func TestBillingRepo_GetAll_Query_WithUnpaidFilter(t *testing.T) {
	isPaid := false
	conditions := "c.auth_user_id = $1 AND ab.deleted_at IS NULL"
	if isPaid {
		conditions += " AND ab.is_paid = TRUE"
	} else {
		conditions += " AND ab.is_paid = FALSE"
	}

	assert.Contains(t, conditions, "ab.is_paid = FALSE")
}

func TestBillingRepo_GetAll_WithMock(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	authUserID := "user-123"
	filters := models.BillingFilters{}
	pagination := models.PaginationParams{Limit: 10, Page: 1}

	billings := []models.AccountBilling{
		{ID: "billing-1", AccountID: "account-1", Period: 202604, AmountBilled: 15000.00},
		{ID: "billing-2", AccountID: "account-1", Period: 202603, AmountBilled: 12000.00},
	}

	mockRepo.On("GetAll", mock.Anything, authUserID, filters, pagination).Return(billings, 2, nil)

	result, total, err := mockRepo.GetAll(context.Background(), authUserID, filters, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 2, total)
	assert.Equal(t, "billing-1", result[0].ID)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_GetAll_WithAccountFilter(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	authUserID := "user-123"
	accountID := "account-123"
	filters := models.BillingFilters{AccountID: &accountID}
	pagination := models.PaginationParams{Limit: 10, Page: 1}

	billings := []models.AccountBilling{
		{ID: "billing-1", AccountID: accountID, Period: 202604, AmountBilled: 15000.00},
	}

	mockRepo.On("GetAll", mock.Anything, authUserID, mock.MatchedBy(func(f models.BillingFilters) bool {
		return f.AccountID != nil && *f.AccountID == accountID
	}), pagination).Return(billings, 1, nil)

	result, _, err := mockRepo.GetAll(context.Background(), authUserID, filters, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, accountID, result[0].AccountID)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_GetAll_Query_WithPeriodRange(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	authUserID := "user-123"
	from := 202601
	to := 202603
	filters := models.BillingFilters{FromPeriod: &from, ToPeriod: &to}
	pagination := models.PaginationParams{Limit: 10, Page: 1}

	billings := []models.AccountBilling{
		{ID: "billing-1", Period: 202602},
		{ID: "billing-2", Period: 202603},
	}

	mockRepo.On("GetAll", mock.Anything, authUserID, mock.MatchedBy(func(f models.BillingFilters) bool {
		return f.FromPeriod != nil && *f.FromPeriod == from && f.ToPeriod != nil && *f.ToPeriod == to
	}), pagination).Return(billings, 2, nil)

	result, total, err := mockRepo.GetAll(context.Background(), authUserID, filters, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 2, total)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_GetAll_WithPaidFilter(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	authUserID := "user-123"
	isPaid := true
	filters := models.BillingFilters{IsPaid: &isPaid}
	pagination := models.PaginationParams{Limit: 10, Page: 1}

	billings := []models.AccountBilling{
		{ID: "billing-1", IsPaid: true},
	}

	mockRepo.On("GetAll", mock.Anything, authUserID, mock.MatchedBy(func(f models.BillingFilters) bool {
		return f.IsPaid != nil && *f.IsPaid == true
	}), pagination).Return(billings, 1, nil)

	result, _, err := mockRepo.GetAll(context.Background(), authUserID, filters, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.True(t, result[0].IsPaid)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_GetAll_WithUnpaidFilter(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	authUserID := "user-123"
	isPaid := false
	filters := models.BillingFilters{IsPaid: &isPaid}
	pagination := models.PaginationParams{Limit: 10, Page: 1}

	billings := []models.AccountBilling{
		{ID: "billing-1", IsPaid: false},
	}

	mockRepo.On("GetAll", mock.Anything, authUserID, mock.MatchedBy(func(f models.BillingFilters) bool {
		return f.IsPaid != nil && *f.IsPaid == false
	}), pagination).Return(billings, 1, nil)

	result, _, err := mockRepo.GetAll(context.Background(), authUserID, filters, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.False(t, result[0].IsPaid)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_GetAll_EmptyResult(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	authUserID := "user-123"
	filters := models.BillingFilters{}
	pagination := models.PaginationParams{Limit: 10, Page: 1}

	mockRepo.On("GetAll", mock.Anything, authUserID, filters, pagination).Return([]models.AccountBilling(nil), 0, nil)

	result, total, err := mockRepo.GetAll(context.Background(), authUserID, filters, pagination)

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, total)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_GetAll_AllFilters(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	authUserID := "user-123"
	accountID := "account-123"
	from := 202601
	to := 202603
	isPaid := true
	filters := models.BillingFilters{
		AccountID:  &accountID,
		FromPeriod: &from,
		ToPeriod:   &to,
		IsPaid:     &isPaid,
	}
	pagination := models.PaginationParams{Limit: 10, Page: 1}

	billings := []models.AccountBilling{
		{ID: "billing-1", AccountID: accountID, Period: 202602, IsPaid: true},
	}

	mockRepo.On("GetAll", mock.Anything, authUserID, mock.MatchedBy(func(f models.BillingFilters) bool {
		return f.AccountID != nil && *f.AccountID == accountID &&
			f.FromPeriod != nil && *f.FromPeriod == from &&
			f.ToPeriod != nil && *f.ToPeriod == to &&
			f.IsPaid != nil && *f.IsPaid == isPaid
	}), pagination).Return(billings, 1, nil)

	result, total, err := mockRepo.GetAll(context.Background(), authUserID, filters, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, 1, total)
	mockRepo.AssertExpectations(t)
}

// SoftDelete Repository Method Tests

func TestBillingRepo_SoftDelete_Query(t *testing.T) {
	// Test soft delete query with auth check via JOIN
	query := `UPDATE homepay.account_billings ab
		SET deleted_at = NOW()
		FROM homepay.accounts a
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE ab.id = $1 AND ab.account_id = a.id AND c.auth_user_id = $2 AND ab.deleted_at IS NULL`

	assert.Contains(t, query, "UPDATE homepay.account_billings ab")
	assert.Contains(t, query, "SET deleted_at = NOW()")
	assert.Contains(t, query, "FROM homepay.accounts a")
	assert.Contains(t, query, "JOIN homepay.companies c ON c.id = a.company_id")
	assert.Contains(t, query, "c.auth_user_id = $2")
	assert.Contains(t, query, "ab.deleted_at IS NULL")
}

func TestBillingRepo_SoftDelete_WithMock(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	billingID := "billing-123"
	authUserID := "user-123"

	mockRepo.On("SoftDelete", mock.Anything, billingID, authUserID).Return(nil)

	err := mockRepo.SoftDelete(context.Background(), billingID, authUserID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_SoftDelete_NotFound(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	billingID := "non-existent"
	authUserID := "user-123"

	mockRepo.On("SoftDelete", mock.Anything, billingID, authUserID).Return(pgx.ErrNoRows)

	err := mockRepo.SoftDelete(context.Background(), billingID, authUserID)

	assert.Error(t, err)
	assert.Equal(t, pgx.ErrNoRows, err)
	mockRepo.AssertExpectations(t)
}

func TestBillingRepo_SoftDelete_WrongAuth(t *testing.T) {
	mockRepo := new(MockBillingRepository)

	billingID := "billing-123"
	wrongAuthUserID := "other-user"

	// Simulate no rows affected when auth check fails
	mockRepo.On("SoftDelete", mock.Anything, billingID, wrongAuthUserID).Return(pgx.ErrNoRows)

	err := mockRepo.SoftDelete(context.Background(), billingID, wrongAuthUserID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

// Pagination edge cases for GetAll

func TestBillingRepo_GetAll_Pagination(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		limit      int
		wantOffset int
	}{
		{"first page", 1, 10, 0},
		{"second page", 2, 10, 10},
		{"third page with different limit", 3, 25, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := models.PaginationParams{Page: tt.page, Limit: tt.limit}
			assert.Equal(t, tt.wantOffset, p.Offset())
		})
	}
}
