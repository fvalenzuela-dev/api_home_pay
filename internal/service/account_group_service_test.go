package service

import (
	"context"
	"testing"

	"github.com/homepay/api/internal/models"
	"github.com/homepay/api/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAccountGroupRepository struct {
	mock.Mock
}

func (m *MockAccountGroupRepository) Create(ctx context.Context, authUserID string, req *models.CreateAccountGroupRequest) (*models.AccountGroup, error) {
	args := m.Called(ctx, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountGroup), args.Error(1)
}

func (m *MockAccountGroupRepository) GetByID(ctx context.Context, id, authUserID string) (*models.AccountGroup, error) {
	args := m.Called(ctx, id, authUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountGroup), args.Error(1)
}

func (m *MockAccountGroupRepository) GetAll(ctx context.Context, authUserID string, p models.PaginationParams) ([]models.AccountGroup, int, error) {
	args := m.Called(ctx, authUserID, p)
	return args.Get(0).([]models.AccountGroup), args.Int(1), args.Error(2)
}

func (m *MockAccountGroupRepository) Update(ctx context.Context, id, authUserID string, req *models.UpdateAccountGroupRequest) (*models.AccountGroup, error) {
	args := m.Called(ctx, id, authUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountGroup), args.Error(1)
}

func (m *MockAccountGroupRepository) SoftDelete(ctx context.Context, id, authUserID string) error {
	args := m.Called(ctx, id, authUserID)
	return args.Error(0)
}

func TestAccountGroupService_Create(t *testing.T) {
	mockRepo := new(MockAccountGroupRepository)
	svc := NewAccountGroupService(mockRepo)

	t.Run("success - create account group", func(t *testing.T) {
		req := &models.CreateAccountGroupRequest{
			Name: "Test Group",
		}
		mockRepo.On("Create", mock.Anything, "user_123", req).Return(&models.AccountGroup{
			ID:         "group-123",
			AuthUserID: "user_123",
			Name:       "Test Group",
		}, nil)

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.NoError(t, err)
		assert.Equal(t, "group-123", result.ID)
		assert.Equal(t, "Test Group", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - name is required", func(t *testing.T) {
		req := &models.CreateAccountGroupRequest{
			Name: "",
		}

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("error - duplicate name", func(t *testing.T) {
		req := &models.CreateAccountGroupRequest{
			Name: "Existing Group",
		}
		mockRepo.On("Create", mock.Anything, "user_123", req).Return(nil, repository.ErrDuplicateName)

		result, err := svc.Create(context.Background(), "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ya existe un grupo con ese nombre")
	})
}

func TestAccountGroupService_GetByID(t *testing.T) {
	mockRepo := new(MockAccountGroupRepository)
	svc := NewAccountGroupService(mockRepo)

	t.Run("success - get account group by ID", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, "group-123", "user_123").Return(&models.AccountGroup{
			ID:         "group-123",
			AuthUserID: "user_123",
			Name:       "Test Group",
		}, nil)

		result, err := svc.GetByID(context.Background(), "group-123", "user_123")

		assert.NoError(t, err)
		assert.Equal(t, "group-123", result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, "group-notfound", "user_123").Return(nil, nil)

		result, err := svc.GetByID(context.Background(), "group-notfound", "user_123")

		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestAccountGroupService_GetAll(t *testing.T) {
	mockRepo := new(MockAccountGroupRepository)
	svc := NewAccountGroupService(mockRepo)

	t.Run("success - get all account groups", func(t *testing.T) {
		groups := []models.AccountGroup{
			{ID: "group-1", Name: "Group 1"},
			{ID: "group-2", Name: "Group 2"},
		}
		mockRepo.On("GetAll", mock.Anything, "user_123", mock.Anything).Return(groups, 2, nil)

		result, total, err := svc.GetAll(context.Background(), "user_123", models.PaginationParams{Page: 1, Limit: 20})

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success - empty list", func(t *testing.T) {
		// Create a fresh mock for this test
		mockRepoEmpty := new(MockAccountGroupRepository)
		svcEmpty := NewAccountGroupService(mockRepoEmpty)
		emptyGroups := []models.AccountGroup{}
		mockRepoEmpty.On("GetAll", mock.Anything, "user_123", mock.Anything).Return(emptyGroups, 0, nil)

		result, total, err := svcEmpty.GetAll(context.Background(), "user_123", models.PaginationParams{Page: 1, Limit: 20})

		assert.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, result)
	})
}

func TestAccountGroupService_Update(t *testing.T) {
	mockRepo := new(MockAccountGroupRepository)
	svc := NewAccountGroupService(mockRepo)

	t.Run("success - update account group", func(t *testing.T) {
		name := "Updated Group"
		req := &models.UpdateAccountGroupRequest{
			Name: &name,
		}
		mockRepo.On("Update", mock.Anything, "group-123", "user_123", req).Return(&models.AccountGroup{
			ID:         "group-123",
			AuthUserID: "user_123",
			Name:       "Updated Group",
		}, nil)

		result, err := svc.Update(context.Background(), "group-123", "user_123", req)

		assert.NoError(t, err)
		assert.Equal(t, "Updated Group", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - name cannot be empty", func(t *testing.T) {
		emptyName := ""
		req := &models.UpdateAccountGroupRequest{
			Name: &emptyName,
		}

		result, err := svc.Update(context.Background(), "group-123", "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})

	t.Run("error - duplicate name", func(t *testing.T) {
		name := "Existing Group"
		req := &models.UpdateAccountGroupRequest{
			Name: &name,
		}
		mockRepo.On("Update", mock.Anything, "group-123", "user_123", req).Return(nil, repository.ErrDuplicateName)

		result, err := svc.Update(context.Background(), "group-123", "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ya existe un grupo con ese nombre")
	})

	t.Run("error - duplicate name", func(t *testing.T) {
		name := "Existing Group"
		req := &models.UpdateAccountGroupRequest{
			Name: &name,
		}
		// When repo returns ErrDuplicateName, service returns error with specific message
		mockRepo.On("Update", mock.Anything, "group-123", "user_123", req).Return(nil, repository.ErrDuplicateName)

		result, err := svc.Update(context.Background(), "group-123", "user_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ya existe un grupo con ese nombre")
	})
}

func TestAccountGroupService_Delete(t *testing.T) {
	mockRepo := new(MockAccountGroupRepository)
	svc := NewAccountGroupService(mockRepo)

	t.Run("success - delete account group", func(t *testing.T) {
		mockRepo.On("SoftDelete", mock.Anything, "group-123", "user_123").Return(nil)

		err := svc.Delete(context.Background(), "group-123", "user_123")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - not found", func(t *testing.T) {
		// Create a fresh mock for this test
		mockRepoDelete := new(MockAccountGroupRepository)
		svcDelete := NewAccountGroupService(mockRepoDelete)
		mockRepoDelete.On("SoftDelete", mock.Anything, "group-123", "user_123").Return(pgx.ErrNoRows)

		err := svcDelete.Delete(context.Background(), "group-123", "user_123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
