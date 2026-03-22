package services

import (
	"testing"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPeriodRepository is a mock implementation of PeriodRepository
type MockPeriodRepository struct {
	mock.Mock
}

func (m *MockPeriodRepository) Create(userID string, period *models.Period) error {
	args := m.Called(userID, period)
	return args.Error(0)
}

func (m *MockPeriodRepository) GetByID(userID string, id int) (*models.Period, error) {
	args := m.Called(userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Period), args.Error(1)
}

func (m *MockPeriodRepository) GetAll(userID string) ([]models.Period, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Period), args.Error(1)
}

func (m *MockPeriodRepository) Update(userID string, period *models.Period) error {
	args := m.Called(userID, period)
	return args.Error(0)
}

func (m *MockPeriodRepository) Delete(userID string, id int) error {
	args := m.Called(userID, id)
	return args.Error(0)
}

func (m *MockPeriodRepository) ExistsByMonthYear(userID string, monthNumber, yearNumber int) (bool, error) {
	args := m.Called(userID, monthNumber, yearNumber)
	return args.Bool(0), args.Error(1)
}

func (m *MockPeriodRepository) HasExpensesOrIncomes(id int) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func TestPeriodService_Create_Success(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	period := &models.Period{MonthNumber: 6, YearNumber: 2024}

	mockRepo.On("ExistsByMonthYear", "user123", 6, 2024).Return(false, nil)
	mockRepo.On("Create", "user123", period).Return(nil)

	err := service.Create("user123", period)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestPeriodService_Create_InvalidMonth(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	period := &models.Period{MonthNumber: 0, YearNumber: 2024}

	err := service.Create("user123", period)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "month must be between 1 and 12")
}

func TestPeriodService_Create_MonthTooHigh(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	period := &models.Period{MonthNumber: 13, YearNumber: 2024}

	err := service.Create("user123", period)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "month must be between 1 and 12")
}

func TestPeriodService_Create_InvalidYear(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	period := &models.Period{MonthNumber: 6, YearNumber: 0}

	err := service.Create("user123", period)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "year must be a positive number")
}

func TestPeriodService_Create_Duplicate(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	period := &models.Period{MonthNumber: 6, YearNumber: 2024}

	mockRepo.On("ExistsByMonthYear", "user123", 6, 2024).Return(true, nil)

	err := service.Create("user123", period)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertExpectations(t)
}

func TestPeriodService_GetByID_Success(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	expectedPeriod := &models.Period{ID: 1, MonthNumber: 6, YearNumber: 2024}

	mockRepo.On("GetByID", "user123", 1).Return(expectedPeriod, nil)

	period, err := service.GetByID("user123", 1)

	assert.NoError(t, err)
	assert.NotNil(t, period)
	assert.Equal(t, 6, period.MonthNumber)
	assert.Equal(t, 2024, period.YearNumber)
	mockRepo.AssertExpectations(t)
}

func TestPeriodService_GetByID_InvalidID(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	period, err := service.GetByID("user123", 0)

	assert.Error(t, err)
	assert.Nil(t, period)
	assert.Contains(t, err.Error(), "invalid period ID")
}

func TestPeriodService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	mockRepo.On("GetByID", "user123", 999).Return(nil, nil)

	period, err := service.GetByID("user123", 999)

	assert.Error(t, err)
	assert.Nil(t, period)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestPeriodService_GetAll_Success(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	expectedPeriods := []models.Period{
		{ID: 1, MonthNumber: 1, YearNumber: 2024},
		{ID: 2, MonthNumber: 2, YearNumber: 2024},
	}

	mockRepo.On("GetAll", "user123").Return(expectedPeriods, nil)

	periods, err := service.GetAll("user123")

	assert.NoError(t, err)
	assert.Len(t, periods, 2)
	mockRepo.AssertExpectations(t)
}

func TestPeriodService_Update_Success(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	existingPeriod := &models.Period{ID: 1, MonthNumber: 6, YearNumber: 2024}
	updatedPeriod := &models.Period{ID: 1, MonthNumber: 7, YearNumber: 2024}

	mockRepo.On("GetByID", "user123", 1).Return(existingPeriod, nil)
	mockRepo.On("ExistsByMonthYear", "user123", 7, 2024).Return(false, nil)
	mockRepo.On("Update", "user123", updatedPeriod).Return(nil)

	err := service.Update("user123", updatedPeriod)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestPeriodService_Update_InvalidID(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	period := &models.Period{ID: 0, MonthNumber: 6, YearNumber: 2024}

	err := service.Update("user123", period)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid period ID")
}

func TestPeriodService_Update_NotFound(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	period := &models.Period{ID: 999, MonthNumber: 6, YearNumber: 2024}

	mockRepo.On("GetByID", "user123", 999).Return(nil, nil)

	err := service.Update("user123", period)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestPeriodService_Update_Duplicate(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	existingPeriod := &models.Period{ID: 1, MonthNumber: 6, YearNumber: 2024}
	updatedPeriod := &models.Period{ID: 1, MonthNumber: 8, YearNumber: 2024}

	mockRepo.On("GetByID", "user123", 1).Return(existingPeriod, nil)
	mockRepo.On("ExistsByMonthYear", "user123", 8, 2024).Return(true, nil)

	err := service.Update("user123", updatedPeriod)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertExpectations(t)
}

func TestPeriodService_Update_SameMonthYear(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	existingPeriod := &models.Period{ID: 1, MonthNumber: 6, YearNumber: 2024}
	updatedPeriod := &models.Period{ID: 1, MonthNumber: 6, YearNumber: 2024}

	mockRepo.On("GetByID", "user123", 1).Return(existingPeriod, nil)
	mockRepo.On("Update", "user123", updatedPeriod).Return(nil)

	err := service.Update("user123", updatedPeriod)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestPeriodService_Delete_Success(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	mockRepo.On("HasExpensesOrIncomes", 1).Return(false, nil)
	mockRepo.On("Delete", "user123", 1).Return(nil)

	err := service.Delete("user123", 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestPeriodService_Delete_InvalidID(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	err := service.Delete("user123", 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid period ID")
}

func TestPeriodService_Delete_HasDependencies(t *testing.T) {
	mockRepo := new(MockPeriodRepository)
	service := NewPeriodService(mockRepo)

	mockRepo.On("HasExpensesOrIncomes", 1).Return(true, nil)

	err := service.Delete("user123", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete period with associated expenses or incomes")
	mockRepo.AssertExpectations(t)
}
