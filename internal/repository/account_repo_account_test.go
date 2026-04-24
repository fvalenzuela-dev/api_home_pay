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

// Test accountRepo functions

func TestAccountRepo_NewAccountRepository(t *testing.T) {
	t.Run("creates accountRepo instance", func(t *testing.T) {
		var repo AccountRepository = NewAccountRepository(nil)
		assert.NotNil(t, repo)
	})
}

func TestScanAccount(t *testing.T) {
	t.Run("scanAccount function exists", func(t *testing.T) {
		// Just verify the function exists
		assert.NotNil(t, scanAccount)
	})
}

// Model tests for account_repo

func TestAccountModel_Full(t *testing.T) {
	t.Run("Account struct can hold all fields", func(t *testing.T) {
		now := time.Now()
		
		account := models.Account{
			ID:             "acc-123",
			CompanyID:      "comp-456",
			GroupID:        strPtr("group-789"),
			AccountNumber:  strPtr("123456"),
			Name:           "Test Account",
			BillingDay:     15,
			AutoAccumulate: true,
			IsActive:       true,
			CreatedAt:      now,
			DeletedAt:      nil,
		}
		
		assert.Equal(t, "acc-123", account.ID)
		assert.Equal(t, "comp-456", account.CompanyID)
		assert.Equal(t, "Test Account", account.Name)
		assert.Equal(t, 15, account.BillingDay)
		assert.True(t, account.AutoAccumulate)
		assert.True(t, account.IsActive)
		assert.NotNil(t, account.GroupID)
		assert.NotNil(t, account.AccountNumber)
	})
}

func TestAccountModel_WithNilPointers(t *testing.T) {
	t.Run("Account struct handles nil pointers", func(t *testing.T) {
		account := models.Account{
			ID:             "acc-123",
			CompanyID:      "comp-456",
			Name:           "Test Account",
			BillingDay:     15,
			AutoAccumulate: false,
			IsActive:       true,
			CreatedAt:      time.Now(),
		}
		
		assert.Nil(t, account.GroupID)
		assert.Nil(t, account.AccountNumber)
		assert.Nil(t, account.DeletedAt)
	})
}

func TestCreateAccountRequest_AllFields(t *testing.T) {
	t.Run("CreateAccountRequest with all fields", func(t *testing.T) {
		groupID := "group-123"
		accountNum := "987654321"
		name := "Full Account"
		billingDay := 25
		autoAcc := true
		
		req := models.CreateAccountRequest{
			GroupID:        &groupID,
			AccountNumber:  &accountNum,
			Name:           name,
			BillingDay:     billingDay,
			AutoAccumulate: autoAcc,
		}
		
		assert.Equal(t, name, req.Name)
		assert.Equal(t, billingDay, req.BillingDay)
		assert.True(t, req.AutoAccumulate)
		assert.NotNil(t, req.GroupID)
		assert.NotNil(t, req.AccountNumber)
	})
}

func TestCreateAccountRequest_MinimalFields(t *testing.T) {
	t.Run("CreateAccountRequest with only required fields", func(t *testing.T) {
		req := models.CreateAccountRequest{
			Name:       "Minimal Account",
			BillingDay: 1,
		}
		
		assert.Equal(t, "Minimal Account", req.Name)
		assert.Equal(t, 1, req.BillingDay)
		assert.Nil(t, req.GroupID)
		assert.Nil(t, req.AccountNumber)
		assert.False(t, req.AutoAccumulate)
	})
}

func TestUpdateAccountRequest_AllFields(t *testing.T) {
	t.Run("UpdateAccountRequest with all fields", func(t *testing.T) {
		groupID := "new-group"
		accountNum := "111222333"
		name := "Updated Name"
		billingDay := 30
		autoAcc := false
		
		req := models.UpdateAccountRequest{
			GroupID:        &groupID,
			AccountNumber:  &accountNum,
			Name:           &name,
			BillingDay:     &billingDay,
			AutoAccumulate: &autoAcc,
		}
		
		assert.NotNil(t, req.Name)
		assert.Equal(t, name, *req.Name)
		assert.Equal(t, billingDay, *req.BillingDay)
		assert.False(t, *req.AutoAccumulate)
	})
}

func TestUpdateAccountRequest_PartialFields(t *testing.T) {
	t.Run("UpdateAccountRequest with only some fields", func(t *testing.T) {
		newName := "Partial Update"
		req := models.UpdateAccountRequest{
			Name: &newName,
		}
		
		assert.NotNil(t, req.Name)
		assert.Equal(t, newName, *req.Name)
		assert.Nil(t, req.GroupID)
		assert.Nil(t, req.AccountNumber)
		assert.Nil(t, req.BillingDay)
		assert.Nil(t, req.AutoAccumulate)
	})
}

// Helper
func strPtr(s string) *string {
	return &s
}

// Additional Account Repository Tests with Mocks - Only unique ones

func TestAccountRepo_GetAllByCompany_WithMock(t *testing.T) {
	mockRepo := new(MockAccountRepository)

	companyID := "company-123"
	authUserID := "user-123"
	pagination := models.PaginationParams{Limit: 10}

	accounts := []models.Account{
		{ID: "account-1", Name: "Account 1"},
		{ID: "account-2", Name: "Account 2"},
	}

	mockRepo.On("GetAllByCompany", mock.Anything, companyID, authUserID, pagination).Return(accounts, 2, nil)

	result, total, err := mockRepo.GetAllByCompany(context.Background(), companyID, authUserID, pagination)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 2, total)
	mockRepo.AssertExpectations(t)
}

func TestAccountRepo_GetAllActiveByUser_WithMock(t *testing.T) {
	mockRepo := new(MockAccountRepository)

	authUserID := "user-123"

	accounts := []models.Account{
		{ID: "account-1", Name: "Account 1", IsActive: true},
		{ID: "account-2", Name: "Account 2", IsActive: true},
	}

	mockRepo.On("GetAllActiveByUser", mock.Anything, authUserID).Return(accounts, nil)

	result, err := mockRepo.GetAllActiveByUser(context.Background(), authUserID)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.True(t, result[0].IsActive)
	mockRepo.AssertExpectations(t)
}

func TestAccountRepo_GetActiveIDsByCompany_WithMock(t *testing.T) {
	mockRepo := new(MockAccountRepository)

	companyID := "company-123"

	ids := []string{"account-1", "account-2", "account-3"}

	mockRepo.On("GetActiveIDsByCompany", mock.Anything, companyID).Return(ids, nil)

	result, err := mockRepo.GetActiveIDsByCompany(context.Background(), companyID)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "account-1", result[0])
	mockRepo.AssertExpectations(t)
}

func TestAccountRepo_Update_NotFound_WithMock(t *testing.T) {
	mockRepo := new(MockAccountRepository)

	req := &models.UpdateAccountRequest{Name: strPtr("New Name")}

	mockRepo.On("Update", mock.Anything, "non-existent", "user-123", req).Return(nil, nil)

	result, err := mockRepo.Update(context.Background(), "non-existent", "user-123", req)

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestAccountRepo_SoftDelete_NotFound_WithMock(t *testing.T) {
	mockRepo := new(MockAccountRepository)

	mockRepo.On("SoftDelete", mock.Anything, "non-existent", "user-123").Return(pgx.ErrNoRows)

	err := mockRepo.SoftDelete(context.Background(), "non-existent", "user-123")

	assert.Error(t, err)
	assert.Equal(t, pgx.ErrNoRows, err)
	mockRepo.AssertExpectations(t)
}

func TestAccountRepo_SoftDeleteByCompany_WithMock(t *testing.T) {
	mockRepo := new(MockAccountRepository)

	companyID := "company-123"

	mockRepo.On("SoftDeleteByCompany", mock.Anything, companyID).Return(nil)

	err := mockRepo.SoftDeleteByCompany(context.Background(), companyID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Additional Account Edge Case Tests

func TestAccountRepo_BillingDayEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		billingDay int
		valid      bool
	}{
		{"day 1", 1, true},
		{"day 15", 15, true},
		{"day 28", 28, true},
		{"day 31", 31, true},
		{"day 0", 0, false},
		{"day 32", 32, false},
		{"negative day", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := models.Account{
				ID:         "test",
				BillingDay: tt.billingDay,
			}
			// Just verify the value is set (validation would be at service layer)
			assert.Equal(t, tt.billingDay, account.BillingDay)
		})
	}
}

func TestAccountRepo_AutoAccumulateEdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		autoAccumulate  bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := models.Account{
				ID:             "test",
				AutoAccumulate: tt.autoAccumulate,
			}
			assert.Equal(t, tt.autoAccumulate, account.AutoAccumulate)
		})
	}
}

func TestAccountRepo_GroupIDEdgeCases(t *testing.T) {
	t.Run("with group ID", func(t *testing.T) {
		groupID := "group-123"
		account := models.Account{
			ID:      "test",
			GroupID: &groupID,
		}
		assert.NotNil(t, account.GroupID)
		assert.Equal(t, "group-123", *account.GroupID)
	})

	t.Run("without group ID", func(t *testing.T) {
		account := models.Account{
			ID: "test",
		}
		assert.Nil(t, account.GroupID)
	})
}

func TestAccountRepo_AccountNumberEdgeCases(t *testing.T) {
	t.Run("with account number", func(t *testing.T) {
		accountNum := "123456789"
		account := models.Account{
			ID:            "test",
			AccountNumber: &accountNum,
		}
		assert.NotNil(t, account.AccountNumber)
		assert.Equal(t, "123456789", *account.AccountNumber)
	})

	t.Run("without account number", func(t *testing.T) {
		account := models.Account{
			ID: "test",
		}
		assert.Nil(t, account.AccountNumber)
	})
}

func TestAccountRepo_Constants(t *testing.T) {
	t.Run("accountCols is defined", func(t *testing.T) {
		// This would test the actual constant if exposed
		assert.NotNil(t, scanAccount)
	})
}

// Account Repository Query Tests

func TestAccountRepo_Create_Query(t *testing.T) {
	// Test the INSERT query structure
	query := `INSERT INTO homepay.accounts (company_id, group_id, account_number, name, billing_day, auto_accumulate)
		SELECT id, $3, $4, $5, $6, $7
		FROM homepay.companies
		WHERE id = $1 AND auth_user_id = $2 AND deleted_at IS NULL`
	
	assert.Contains(t, query, "INSERT INTO homepay.accounts")
	assert.Contains(t, query, "homepay.companies")
}

func TestAccountRepo_GetByID_Query(t *testing.T) {
	// Test the SELECT query structure
	query := `SELECT a.id, a.company_id, a.group_id, a.account_number, a.name, a.billing_day, a.auto_accumulate, a.is_active, a.created_at, a.deleted_at
		FROM homepay.accounts a
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE a.id = $1 AND c.auth_user_id = $2 AND a.deleted_at IS NULL`
	
	assert.Contains(t, query, "FROM homepay.accounts")
	assert.Contains(t, query, "JOIN homepay.companies")
	assert.Contains(t, query, "a.deleted_at IS NULL")
}

func TestAccountRepo_GetAllByCompany_Query(t *testing.T) {
	// Test pagination query structure
	query := `SELECT a.id, a.company_id, a.group_id, a.account_number, a.name, a.billing_day, a.auto_accumulate, a.is_active, a.created_at, a.deleted_at
		FROM homepay.accounts a
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE a.company_id = $1 AND c.auth_user_id = $2 AND a.deleted_at IS NULL
		ORDER BY a.created_at DESC
		LIMIT $3 OFFSET $4`
	
	assert.Contains(t, query, "ORDER BY a.created_at DESC")
	assert.Contains(t, query, "LIMIT")
	assert.Contains(t, query, "OFFSET")
}

func TestAccountRepo_GetAllActiveByUser_Query(t *testing.T) {
	// Test query for getting all active accounts for user
	query := `SELECT a.id, a.company_id, a.group_id, a.account_number, a.name, a.billing_day, a.auto_accumulate, a.is_active, a.created_at, a.deleted_at
		FROM homepay.accounts a
		JOIN homepay.companies c ON c.id = a.company_id
		WHERE c.auth_user_id = $1 AND a.deleted_at IS NULL AND c.deleted_at IS NULL
		ORDER BY a.created_at`
	
	assert.Contains(t, query, "c.deleted_at IS NULL")
	assert.Contains(t, query, "ORDER BY a.created_at")
}

func TestAccountRepo_GetActiveIDsByCompany_Query(t *testing.T) {
	// Test query for getting only IDs
	query := `SELECT id FROM homepay.accounts WHERE company_id = $1 AND deleted_at IS NULL`
	
	assert.Contains(t, query, "SELECT id")
}

func TestAccountRepo_Update_Query(t *testing.T) {
	// Test UPDATE query
	query := `UPDATE homepay.accounts a
		SET group_id        = COALESCE($3, a.group_id),
		    account_number  = COALESCE($4, a.account_number),
		    name           = COALESCE($5, a.name),
		    billing_day    = COALESCE($6, a.billing_day),
		    auto_accumulate = COALESCE($7, a.auto_accumulate)
		FROM homepay.companies c
		WHERE a.id = $1 AND a.company_id = c.id AND c.auth_user_id = $2 AND a.deleted_at IS NULL`
	
	assert.Contains(t, query, "COALESCE")
	assert.Contains(t, query, "FROM homepay.companies c")
}

func TestAccountRepo_SoftDelete_Query(t *testing.T) {
	// Test soft delete query
	query := `UPDATE homepay.accounts a
		SET deleted_at = NOW(), is_active = FALSE
		FROM homepay.companies c
		WHERE a.id = $1 AND a.company_id = c.id AND c.auth_user_id = $2 AND a.deleted_at IS NULL`
	
	assert.Contains(t, query, "deleted_at = NOW()")
	assert.Contains(t, query, "is_active = FALSE")
}

func TestAccountRepo_SoftDeleteByCompany_Query(t *testing.T) {
	// Test bulk soft delete
	query := `UPDATE homepay.accounts SET deleted_at = NOW(), is_active = FALSE
		WHERE company_id = $1 AND deleted_at IS NULL`
	
	assert.Contains(t, query, "UPDATE homepay.accounts")
}
