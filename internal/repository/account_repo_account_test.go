package repository

import (
	"testing"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
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
