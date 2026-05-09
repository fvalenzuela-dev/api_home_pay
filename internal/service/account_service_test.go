package service

import (
	"context"
	"errors"
	"testing"

	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAccountService_Create(t *testing.T) {
	mockAccounts := new(AccountRepoMock)
	mockCompanies := new(CompanyRepoMock)
	mockBillings := new(BillingRepoMock)
	svc := NewAccountService(mockAccounts, mockCompanies, mockBillings)

	t.Run("success - create account", func(t *testing.T) {
		req := &models.CreateAccountRequest{
			CompanyID:      "company-123",
			Name:           "Test Account",
			BillingDay:     15,
			AutoAccumulate: true,
		}
		mockCompanies.On("GetByID", mock.Anything, "company-123", "user_123").Return(&models.Company{
			ID: "company-123",
		}, nil)
		mockAccounts.On("Create", mock.Anything, "company-123", "user_123", mock.Anything).Return(&models.Account{
			ID:         "account-123",
			CompanyID:  "company-123",
			Name:       "Test Account",
			BillingDay: 15,
		}, nil)

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.NoError(t, err)
		assert.Equal(t, "account-123", result.ID)
		mockAccounts.AssertExpectations(t)
	})

	t.Run("error - empty name", func(t *testing.T) {
		req := &models.CreateAccountRequest{
			CompanyID:  "company-123",
			Name:       "",
			BillingDay: 15,
		}

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("error - invalid billing day", func(t *testing.T) {
		req := &models.CreateAccountRequest{
			CompanyID:  "company-123",
			Name:       "Test Account",
			BillingDay: 0,
		}

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "billing_day must be between 1 and 31")
	})

	t.Run("error - billing day too high", func(t *testing.T) {
		req := &models.CreateAccountRequest{
			CompanyID:  "company-123",
			Name:       "Test Account",
			BillingDay: 32,
		}

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "billing_day must be between 1 and 31")
	})

	t.Run("error - company not found", func(t *testing.T) {
		req := &models.CreateAccountRequest{
			CompanyID:  "company-123",
			Name:       "Test Account",
			BillingDay: 15,
		}
		mockCompanies.On("GetByID", mock.Anything, "company-123", "user_123").Return(nil, nil).Maybe()
		mockAccounts.On("Create", mock.Anything, "company-123", "user_123", mock.Anything).Return(nil, errors.New("should not be called")).Maybe()

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "company not found or access denied")
	})
}

func TestAccountService_GetByID(t *testing.T) {
	mockAccounts := new(AccountRepoMock)
	mockCompanies := new(CompanyRepoMock)
	mockBillings := new(BillingRepoMock)
	svc := NewAccountService(mockAccounts, mockCompanies, mockBillings)

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

func TestAccountService_GetAll(t *testing.T) {
	mockAccounts := new(AccountRepoMock)
	mockCompanies := new(CompanyRepoMock)
	mockBillings := new(BillingRepoMock)
	svc := NewAccountService(mockAccounts, mockCompanies, mockBillings)

	t.Run("success - list all accounts", func(t *testing.T) {
		accounts := []models.Account{
			{ID: "account-1", Name: "Account 1"},
			{ID: "account-2", Name: "Account 2"},
		}
		mockAccounts.On("GetAllFiltered", mock.Anything, "user_123", mock.Anything, "", "", mock.Anything).Return(accounts, 2, nil)

		result, total, err := svc.GetAll(context.Background(), "user_123", nil, "", "", models.PaginationParams{Page: 1, Limit: 20})

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, result, 2)
		mockAccounts.AssertExpectations(t)
	})

	t.Run("success - filter by company", func(t *testing.T) {
		accounts := []models.Account{
			{ID: "account-1", Name: "Account 1"},
		}
		companyID := "company-123"
		mockAccounts.On("GetAllFiltered", mock.Anything, "user_123", &companyID, "", "", mock.Anything).Return(accounts, 1, nil)

		result, total, err := svc.GetAll(context.Background(), "user_123", &companyID, "", "", models.PaginationParams{Page: 1, Limit: 20})

		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, result, 1)
		mockAccounts.AssertExpectations(t)
	})

	t.Run("success - with sort and order", func(t *testing.T) {
		accounts := []models.Account{
			{ID: "account-1", Name: "Account 1"},
		}
		mockAccounts.On("GetAllFiltered", mock.Anything, "user_123", mock.Anything, "name", "asc", mock.Anything).Return(accounts, 1, nil)

		result, total, err := svc.GetAll(context.Background(), "user_123", nil, "name", "asc", models.PaginationParams{Page: 1, Limit: 20})

		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, result, 1)
		mockAccounts.AssertExpectations(t)
	})
}

func TestAccountService_Update(t *testing.T) {
	mockAccounts := new(AccountRepoMock)
	mockCompanies := new(CompanyRepoMock)
	mockBillings := new(BillingRepoMock)
	svc := NewAccountService(mockAccounts, mockCompanies, mockBillings)

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
	mockCompanies := new(CompanyRepoMock)
	mockBillings := new(BillingRepoMock)
	svc := NewAccountService(mockAccounts, mockCompanies, mockBillings)

	t.Run("success", func(t *testing.T) {
		mockBillings.On("SoftDeleteByAccount", mock.Anything, "account-123").Return(nil)
		mockAccounts.On("SoftDelete", mock.Anything, "account-123", "user_123").Return(nil)

		err := svc.Delete(context.Background(), "account-123", "user_123")

		assert.NoError(t, err)
		mockBillings.AssertExpectations(t)
		mockAccounts.AssertExpectations(t)
	})
}
