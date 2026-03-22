package services

import (
	"testing"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockIncomeRepository is a mock implementation of IncomeRepository
type MockIncomeRepository struct {
	mock.Mock
}

func (m *MockIncomeRepository) Create(userID string, income *models.Income) error {
	args := m.Called(userID, income)
	return args.Error(0)
}

func (m *MockIncomeRepository) GetByID(userID string, id int) (*models.Income, error) {
	args := m.Called(userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Income), args.Error(1)
}

func (m *MockIncomeRepository) GetAll(userID string, periodID *int) ([]models.Income, error) {
	args := m.Called(userID, periodID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Income), args.Error(1)
}

func (m *MockIncomeRepository) Update(userID string, income *models.Income) error {
	args := m.Called(userID, income)
	return args.Error(0)
}

func (m *MockIncomeRepository) Delete(userID string, id int) error {
	args := m.Called(userID, id)
	return args.Error(0)
}

func (m *MockIncomeRepository) PeriodExistsAndBelongsToUser(userID string, periodID int) (bool, error) {
	args := m.Called(userID, periodID)
	return args.Bool(0), args.Error(1)
}

func (m *MockIncomeRepository) GetTotalByPeriod(userID string, periodID int) (float64, int, error) {
	args := m.Called(userID, periodID)
	return args.Get(0).(float64), args.Int(1), args.Error(2)
}

func TestIncomeService_Create_Success(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	income := &models.Income{
		PeriodID:    1,
		Description: "Salary",
		Amount:      5000.00,
		IsRecurring: true,
		ReceivedAt:  "2024-06-01",
	}

	mockRepo.On("PeriodExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("Create", "user123", income).Return(nil)

	err := service.Create("user123", income)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestIncomeService_Create_EmptyDescription(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	income := &models.Income{
		PeriodID:    1,
		Description: "",
		Amount:      5000.00,
	}

	err := service.Create("user123", income)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description cannot be empty")
}

func TestIncomeService_Create_WhitespaceDescription(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	income := &models.Income{
		PeriodID:    1,
		Description: "   ",
		Amount:      5000.00,
	}

	err := service.Create("user123", income)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description cannot be empty")
}

func TestIncomeService_Create_ZeroAmount(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	income := &models.Income{
		PeriodID:    1,
		Description: "Salary",
		Amount:      0,
	}

	err := service.Create("user123", income)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be greater than zero")
}

func TestIncomeService_Create_NegativeAmount(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	income := &models.Income{
		PeriodID:    1,
		Description: "Salary",
		Amount:      -1000.00,
	}

	err := service.Create("user123", income)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be greater than zero")
}

func TestIncomeService_Create_InvalidPeriodID(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	income := &models.Income{
		PeriodID:    0,
		Description: "Salary",
		Amount:      5000.00,
	}

	mockRepo.On("PeriodExistsAndBelongsToUser", "user123", 0).Return(false, nil)

	err := service.Create("user123", income)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "period not found or access denied")
	mockRepo.AssertExpectations(t)
}

func TestIncomeService_Create_InvalidReceivedDate(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	income := &models.Income{
		PeriodID:    1,
		Description: "Salary",
		Amount:      5000.00,
		ReceivedAt:  "invalid-date",
	}

	err := service.Create("user123", income)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid received_at date format")
}

func TestIncomeService_Create_PeriodNotFound(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	income := &models.Income{
		PeriodID:    999,
		Description: "Salary",
		Amount:      5000.00,
	}

	mockRepo.On("PeriodExistsAndBelongsToUser", "user123", 999).Return(false, nil)

	err := service.Create("user123", income)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "period not found or access denied")
	mockRepo.AssertExpectations(t)
}

func TestIncomeService_GetByID_Success(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	expectedIncome := &models.Income{
		ID:          1,
		PeriodID:    1,
		Description: "Salary",
		Amount:      5000.00,
	}

	mockRepo.On("GetByID", "user123", 1).Return(expectedIncome, nil)

	income, err := service.GetByID("user123", 1)

	assert.NoError(t, err)
	assert.NotNil(t, income)
	assert.Equal(t, "Salary", income.Description)
	assert.Equal(t, 5000.00, income.Amount)
	mockRepo.AssertExpectations(t)
}

func TestIncomeService_GetByID_InvalidID(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	income, err := service.GetByID("user123", 0)

	assert.Error(t, err)
	assert.Nil(t, income)
	assert.Contains(t, err.Error(), "invalid income ID")
}

func TestIncomeService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	mockRepo.On("GetByID", "user123", 999).Return(nil, nil)

	income, err := service.GetByID("user123", 999)

	assert.Error(t, err)
	assert.Nil(t, income)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestIncomeService_GetAll_Success(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	expectedIncomes := []models.Income{
		{ID: 1, Description: "Salary"},
		{ID: 2, Description: "Bonus"},
	}

	mockRepo.On("GetAll", "user123", (*int)(nil)).Return(expectedIncomes, nil)

	incomes, err := service.GetAll("user123", nil)

	assert.NoError(t, err)
	assert.Len(t, incomes, 2)
	mockRepo.AssertExpectations(t)
}

func TestIncomeService_GetAll_WithPeriodFilter(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	periodID := 1

	mockRepo.On("PeriodExistsAndBelongsToUser", "user123", 1).Return(true, nil)

	expectedIncomes := []models.Income{
		{ID: 1, PeriodID: 1, Description: "Salary"},
	}

	mockRepo.On("GetAll", "user123", &periodID).Return(expectedIncomes, nil)

	incomes, err := service.GetAll("user123", &periodID)

	assert.NoError(t, err)
	assert.Len(t, incomes, 1)
	mockRepo.AssertExpectations(t)
}

func TestIncomeService_GetAll_WithPeriodFilterNotFound(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	periodID := 999

	mockRepo.On("PeriodExistsAndBelongsToUser", "user123", 999).Return(false, nil)

	incomes, err := service.GetAll("user123", &periodID)

	assert.Error(t, err)
	assert.Nil(t, incomes)
	assert.Contains(t, err.Error(), "period not found or access denied")
	mockRepo.AssertExpectations(t)
}

func TestIncomeService_Update_Success(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	existingIncome := &models.Income{
		ID:          1,
		PeriodID:    1,
		Description: "Old Salary",
		Amount:      5000.00,
	}
	updatedIncome := &models.Income{
		ID:          1,
		PeriodID:    1,
		Description: "New Salary",
		Amount:      5500.00,
	}

	mockRepo.On("GetByID", "user123", 1).Return(existingIncome, nil)
	mockRepo.On("PeriodExistsAndBelongsToUser", "user123", 1).Return(true, nil)
	mockRepo.On("Update", "user123", updatedIncome).Return(nil)

	err := service.Update("user123", updatedIncome)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestIncomeService_Update_InvalidID(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	income := &models.Income{
		ID:          0,
		Description: "Test",
		Amount:      1000.00,
	}

	err := service.Update("user123", income)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid income ID")
}

func TestIncomeService_Update_EmptyDescription(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	income := &models.Income{
		ID:          1,
		Description: "",
		Amount:      1000.00,
	}

	err := service.Update("user123", income)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description cannot be empty")
}

func TestIncomeService_Update_ZeroAmount(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	income := &models.Income{
		ID:          1,
		Description: "Test",
		Amount:      0,
	}

	err := service.Update("user123", income)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be greater than zero")
}

func TestIncomeService_Update_NotFound(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	income := &models.Income{
		ID:          999,
		Description: "Test",
		Amount:      1000.00,
	}

	mockRepo.On("GetByID", "user123", 999).Return(nil, nil)

	err := service.Update("user123", income)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestIncomeService_Update_PeriodNotFound(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	existingIncome := &models.Income{
		ID:          1,
		PeriodID:    1,
		Description: "Old",
		Amount:      5000.00,
	}
	updatedIncome := &models.Income{
		ID:          1,
		PeriodID:    999,
		Description: "New",
		Amount:      5500.00,
	}

	mockRepo.On("GetByID", "user123", 1).Return(existingIncome, nil)
	mockRepo.On("PeriodExistsAndBelongsToUser", "user123", 999).Return(false, nil)

	err := service.Update("user123", updatedIncome)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "period not found or access denied")
	mockRepo.AssertExpectations(t)
}

func TestIncomeService_Update_InvalidReceivedDate(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	updatedIncome := &models.Income{
		ID:          1,
		PeriodID:    1,
		Description: "Salary",
		Amount:      5500.00,
		ReceivedAt:  "invalid-date",
	}

	err := service.Update("user123", updatedIncome)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid received_at date format")
}

func TestIncomeService_Delete_Success(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	existingIncome := &models.Income{
		ID:          1,
		PeriodID:    1,
		Description: "Salary",
		Amount:      5000.00,
	}

	mockRepo.On("GetByID", "user123", 1).Return(existingIncome, nil)
	mockRepo.On("Delete", "user123", 1).Return(nil)

	err := service.Delete("user123", 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestIncomeService_Delete_InvalidID(t *testing.T) {
	mockRepo := new(MockIncomeRepository)
	service := NewIncomeService(mockRepo)

	err := service.Delete("user123", 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid income ID")
}
