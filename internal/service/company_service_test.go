package service

import (
	"context"
	"testing"

	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type CompanyRepoMock struct {
	mock.Mock
}

func (m *CompanyRepoMock) Create(ctx context.Context, authUserID string, req *models.CreateCompanyRequest) (*models.Company, error) {
	args := m.Called(ctx, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *CompanyRepoMock) GetByID(ctx context.Context, id, authUserID string) (*models.Company, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *CompanyRepoMock) GetAll(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.Company, int, error) {
	args := m.Called(ctx, authUserID, p)
	return args.Get(0).([]models.Company), args.Int(1), args.Error(2)
}

func (m *CompanyRepoMock) Update(ctx context.Context, id, authUserID string, req *models.UpdateCompanyRequest) (*models.Company, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *CompanyRepoMock) SoftDelete(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

type AccountRepoMock struct {
	mock.Mock
}

func (m *AccountRepoMock) Create(ctx context.Context, companyID, authUserID string, req *models.CreateAccountRequest) (*models.Account, error) {
	args := m.Called(ctx, companyID, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *AccountRepoMock) GetByID(ctx context.Context, id, authUserID string) (*models.Account, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *AccountRepoMock) GetAllFiltered(ctx context.Context, authUserID string, companyID *string, sort, order string, p models.PaginationParams) ([]models.Account, int, error) {
	args := m.Called(ctx, authUserID, companyID, sort, order, p)
	return args.Get(0).([]models.Account), args.Int(1), args.Error(2)
}

func (m *AccountRepoMock) GetAllActiveByUser(ctx context.Context, authUserID string) ([]models.Account, error) {
	args := m.Called(ctx, authUserID)
	return args.Get(0).([]models.Account), args.Error(1)
}

func (m *AccountRepoMock) GetActiveIDsByCompany(ctx context.Context, companyID string) ([]string, error) {
	args := m.Called(ctx, companyID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *AccountRepoMock) Update(ctx context.Context, id, authUserID string, req *models.UpdateAccountRequest) (*models.Account, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Account), args.Error(1)
}

func (m *AccountRepoMock) SoftDelete(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

func (m *AccountRepoMock) SoftDeleteByCompany(ctx context.Context, companyID string) error {
	args := m.Called(ctx, companyID)
	return args.Error(0)
}

type BillingRepoMock struct {
	mock.Mock
}

func (m *BillingRepoMock) Create(ctx context.Context, accountID string, req *models.CreateBillingRequest) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *BillingRepoMock) CreateCarryOver(ctx context.Context, accountID string, period int, amount float64, carriedFrom string) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID, period, amount, carriedFrom)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *BillingRepoMock) GetByID(ctx context.Context, id, authUserID string) (*models.AccountBilling, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *BillingRepoMock) GetByAccountAndPeriod(ctx context.Context, accountID, authUserID string, period int) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID, authUserID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *BillingRepoMock) GetAllByAccount(ctx context.Context, accountID, authUserID string, p models.PaginationParams) ([]models.AccountBilling, int, error) {
	args := m.Called(ctx, accountID, authUserID, p)
	return args.Get(0).([]models.AccountBilling), args.Int(1), args.Error(2)
}

func (m *BillingRepoMock) GetUnpaidByAccount(ctx context.Context, accountID, authUserID string) (*models.AccountBilling, error) {
	args := m.Called(ctx, accountID, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *BillingRepoMock) GetAllByPeriod(ctx context.Context, authUserID string, period int, isPaid *bool, p models.PaginationParams) ([]models.AccountBillingWithDetails, int, error) {
	args := m.Called(ctx, authUserID, period, isPaid, p)
	return args.Get(0).([]models.AccountBillingWithDetails), args.Int(1), args.Error(2)
}

func (m *BillingRepoMock) BulkInsertForPeriod(ctx context.Context, period int, inserts []models.PeriodBillingInsert) error {
	args := m.Called(ctx, period, inserts)
	return args.Error(0)
}

func (m *BillingRepoMock) Update(ctx context.Context, id, authUserID string, req *models.UpdateBillingRequest) (*models.AccountBilling, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountBilling), args.Error(1)
}

func (m *BillingRepoMock) MarkPaid(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *BillingRepoMock) SoftDeleteByAccount(ctx context.Context, accountID string) error {
	args := m.Called(ctx, accountID)
	return args.Error(0)
}

func TestCompanyService_Create(t *testing.T) {
	mockCompanies := new(CompanyRepoMock)
	mockAccounts := new(AccountRepoMock)
	mockBillings := new(BillingRepoMock)
	svc := NewCompanyService(mockCompanies, mockAccounts, mockBillings)

	t.Run("success - create company", func(t *testing.T) {
		req := &models.CreateCompanyRequest{
			Name:       "Test Company",
			CategoryID: 1,
		}
		mockCompanies.On("Create", mock.Anything, "user_123", req).Return(&models.Company{
			ID:         "company-123",
			AuthUserID: "user_123",
			Name:       "Test Company",
			CategoryID: 1,
		}, nil)

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.NoError(t, err)
		assert.Equal(t, "company-123", result.ID)
		mockCompanies.AssertExpectations(t)
	})

	t.Run("error - empty name", func(t *testing.T) {
		req := &models.CreateCompanyRequest{
			Name:       "",
			CategoryID: 1,
		}

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "name is required")
	})
}

func TestCompanyService_GetByID(t *testing.T) {
	mockCompanies := new(CompanyRepoMock)
	mockAccounts := new(AccountRepoMock)
	mockBillings := new(BillingRepoMock)
	svc := NewCompanyService(mockCompanies, mockAccounts, mockBillings)

	t.Run("success", func(t *testing.T) {
		mockCompanies.On("GetByID", mock.Anything, "company-123", "user_123").Return(&models.Company{
			ID:         "company-123",
			AuthUserID: "user_123",
			Name:       "Test Company",
		}, nil)

		result, err := svc.GetByID(context.Background(), "company-123", "user_123")

		assert.NoError(t, err)
		assert.Equal(t, "company-123", result.ID)
		mockCompanies.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockCompanies.On("GetByID", mock.Anything, "company-notfound", "user_123").Return(nil, nil)

		result, err := svc.GetByID(context.Background(), "company-notfound", "user_123")

		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestCompanyService_GetAll(t *testing.T) {
	mockCompanies := new(CompanyRepoMock)
	mockAccounts := new(AccountRepoMock)
	mockBillings := new(BillingRepoMock)
	svc := NewCompanyService(mockCompanies, mockAccounts, mockBillings)

	t.Run("success", func(t *testing.T) {
		companies := []models.Company{
			{ID: "company-1", Name: "Company 1"},
			{ID: "company-2", Name: "Company 2"},
		}
		mockCompanies.On("GetAll", mock.Anything, "user_123", mock.Anything).Return(companies, 2, nil)

		result, total, err := svc.GetAll(context.Background(), "user_123", models.PaginationParams{Page: 1, Limit: 20})

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, result, 2)
		mockCompanies.AssertExpectations(t)
	})
}

func TestCompanyService_Update(t *testing.T) {
	mockCompanies := new(CompanyRepoMock)
	mockAccounts := new(AccountRepoMock)
	mockBillings := new(BillingRepoMock)
	svc := NewCompanyService(mockCompanies, mockAccounts, mockBillings)

	t.Run("success", func(t *testing.T) {
		name := "Updated Company"
		req := &models.UpdateCompanyRequest{Name: &name}
		mockCompanies.On("Update", mock.Anything, "company-123", "user_123", req).Return(&models.Company{
			ID:   "company-123",
			Name: "Updated Company",
		}, nil)

		result, err := svc.Update(context.Background(), "company-123", "user_123", req)

		assert.NoError(t, err)
		assert.Equal(t, "Updated Company", result.Name)
		mockCompanies.AssertExpectations(t)
	})
}

func TestCompanyService_Delete(t *testing.T) {
	mockCompanies := new(CompanyRepoMock)
	mockAccounts := new(AccountRepoMock)
	mockBillings := new(BillingRepoMock)
	svc := NewCompanyService(mockCompanies, mockAccounts, mockBillings)

	t.Run("success - delete with accounts", func(t *testing.T) {
		mockAccounts.On("GetActiveIDsByCompany", mock.Anything, "company-123").Return([]string{"acc-1", "acc-2"}, nil)
		mockBillings.On("SoftDeleteByAccount", mock.Anything, "acc-1").Return(nil)
		mockBillings.On("SoftDeleteByAccount", mock.Anything, "acc-2").Return(nil)
		mockAccounts.On("SoftDeleteByCompany", mock.Anything, "company-123").Return(nil)
		mockCompanies.On("SoftDelete", mock.Anything, "company-123", "user_123").Return(nil)

		err := svc.Delete(context.Background(), "company-123", "user_123")

		assert.NoError(t, err)
		mockAccounts.AssertExpectations(t)
		mockBillings.AssertExpectations(t)
		mockCompanies.AssertExpectations(t)
	})

	t.Run("success - delete without accounts", func(t *testing.T) {
		mockAccounts.On("GetActiveIDsByCompany", mock.Anything, "company-123").Return([]string{}, nil)
		mockAccounts.On("SoftDeleteByCompany", mock.Anything, "company-123").Return(nil)
		mockCompanies.On("SoftDelete", mock.Anything, "company-123", "user_123").Return(nil)

		err := svc.Delete(context.Background(), "company-123", "user_123")

		assert.NoError(t, err)
		mockAccounts.AssertExpectations(t)
		mockCompanies.AssertExpectations(t)
	})
}
