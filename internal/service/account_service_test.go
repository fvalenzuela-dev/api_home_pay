package service

import (
	"context"
	"testing"

	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAccountService_Create(t *testing.T) {
	mockAccounts := new(AccountRepoMock)
	mockBillings := new(BillingRepoMock)
	svc := NewAccountService(mockAccounts, mockBillings)

	t.Run("success - create account", func(t *testing.T) {
		req := &models.CreateAccountRequest{
			Name:            "Test Account",
			BillingDay:      15,
			AutoAccumulate: true,
		}
		mockAccounts.On("Create", mock.Anything, "company-123", "user_123", req).Return(&models.Account{
			ID:         "account-123",
			CompanyID:  "company-123",
			Name:       "Test Account",
			BillingDay: 15,
		}, nil)

		result, err := svc.Create(context.Background(), "company-123", "user_123", req)

		assert.NoError(t, err)
		assert.Equal(t, "account-123", result.ID)
		mockAccounts.AssertExpectations(t)
	})

	t.Run("error - empty name", func(t *testing.T) {
		req := &models.CreateAccountRequest{
			Name:       "",
			BillingDay: 15,
		}

		result, err := svc.Create(context.Background(), "company-123", "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("error - invalid billing day", func(t *testing.T) {
		req := &models.CreateAccountRequest{
			Name:       "Test Account",
			BillingDay: 0,
		}

		result, err := svc.Create(context.Background(), "company-123", "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "billing_day must be between 1 and 31")
	})

	t.Run("error - billing day too high", func(t *testing.T) {
		req := &models.CreateAccountRequest{
			Name:       "Test Account",
			BillingDay: 32,
		}

		result, err := svc.Create(context.Background(), "company-123", "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "billing_day must be between 1 and 31")
	})
}

func TestAccountService_GetByID(t *testing.T) {
	mockAccounts := new(AccountRepoMock)
	mockBillings := new(BillingRepoMock)
	svc := NewAccountService(mockAccounts, mockBillings)

	t.Run("success", func(t *testing.T) {
		mockAccounts.On("GetByID", mock.Anything, "account-123", "user_123").Return(&models.Account{
			ID:         "account-123",
			CompanyID:  "company-123",
			Name:       "Test Account",
		}, nil)

		result, err := svc.GetByID(context.Background(), "account-123", "user_123")

		assert.NoError(t, err)
		assert.Equal(t, "account-123", result.ID)
		mockAccounts.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockAccounts.On("GetByID", mock.Anything, "account-notfound", "user_123").Return(nil, nil)

		result, err := svc.GetByID(context.Background(), "account-notfound", "user_123")

		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestAccountService_GetAllByCompany(t *testing.T) {
	mockAccounts := new(AccountRepoMock)
	mockBillings := new(BillingRepoMock)
	svc := NewAccountService(mockAccounts, mockBillings)

	t.Run("success", func(t *testing.T) {
		accounts := []models.Account{
			{ID: "account-1", Name: "Account 1"},
			{ID: "account-2", Name: "Account 2"},
		}
		mockAccounts.On("GetAllByCompany", mock.Anything, "company-123", "user_123", mock.Anything).Return(accounts, 2, nil)

		result, total, err := svc.GetAllByCompany(context.Background(), "company-123", "user_123", models.PaginationParams{Page: 1, Limit: 20})

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, result, 2)
		mockAccounts.AssertExpectations(t)
	})
}

func TestAccountService_Update(t *testing.T) {
	mockAccounts := new(AccountRepoMock)
	mockBillings := new(BillingRepoMock)
	svc := NewAccountService(mockAccounts, mockBillings)

	t.Run("success", func(t *testing.T) {
		name := "Updated Account"
		req := &models.UpdateAccountRequest{Name: &name}
		mockAccounts.On("Update", mock.Anything, "account-123", "user_123", req).Return(&models.Account{
			ID:   "account-123",
			Name: "Updated Account",
		}, nil)

		result, err := svc.Update(context.Background(), "account-123", "user_123", req)

		assert.NoError(t, err)
		assert.Equal(t, "Updated Account", result.Name)
		mockAccounts.AssertExpectations(t)
	})

	t.Run("error - invalid billing day", func(t *testing.T) {
		billingDay := 0
		req := &models.UpdateAccountRequest{BillingDay: &billingDay}

		result, err := svc.Update(context.Background(), "account-123", "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "billing_day must be between 1 and 31")
	})

	t.Run("error - billing day too high", func(t *testing.T) {
		billingDay := 32
		req := &models.UpdateAccountRequest{BillingDay: &billingDay}

		result, err := svc.Update(context.Background(), "account-123", "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "billing_day must be between 1 and 31")
	})
}

func TestAccountService_Delete(t *testing.T) {
	mockAccounts := new(AccountRepoMock)
	mockBillings := new(BillingRepoMock)
	svc := NewAccountService(mockAccounts, mockBillings)

	t.Run("success", func(t *testing.T) {
		mockBillings.On("SoftDeleteByAccount", mock.Anything, "account-123").Return(nil)
		mockAccounts.On("SoftDelete", mock.Anything, "account-123", "user_123").Return(nil)

		err := svc.Delete(context.Background(), "account-123", "user_123")

		assert.NoError(t, err)
		mockBillings.AssertExpectations(t)
		mockAccounts.AssertExpectations(t)
	})
}
