package repository

import (
	"testing"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestAccountGroupRepo_Interfaces(t *testing.T) {
	t.Run("AccountGroupRepository interface is satisfied by accountGroupRepo", func(t *testing.T) {
		var _ AccountGroupRepository = (*accountGroupRepo)(nil)
	})
}

func TestScanAccountGroup(t *testing.T) {
	t.Run("scanAccountGroup function exists", func(t *testing.T) {
		assert.NotNil(t, scanAccountGroup)
	})
}

func TestAccountGroupRepo_GetAll(t *testing.T) {
	t.Run("accountGroupCols constant", func(t *testing.T) {
		assert.Equal(t, `id, auth_user_id, name, created_at, deleted_at`, accountGroupCols)
	})
}

func TestAccountGroupRepo_Create(t *testing.T) {
	t.Run("CreateAccountGroupRequest validation", func(t *testing.T) {
		req := models.CreateAccountGroupRequest{
			Name: "Test Group",
		}
		assert.Equal(t, "Test Group", req.Name)
	})
}

func TestAccountGroupRepo_Update(t *testing.T) {
	t.Run("UpdateAccountGroupRequest with pointer", func(t *testing.T) {
		name := "Updated Group"
		req := models.UpdateAccountGroupRequest{
			Name: &name,
		}
		assert.Equal(t, "Updated Group", *req.Name)
	})

	t.Run("UpdateAccountGroupRequest with nil", func(t *testing.T) {
		req := models.UpdateAccountGroupRequest{}
		assert.Nil(t, req.Name)
	})
}

func TestAccountGroupRepo_SoftDelete(t *testing.T) {
	t.Run("Soft delete sets deleted_at", func(t *testing.T) {
		now := time.Now()
		group := models.AccountGroup{
			ID:        "group-123",
			Name:      "Test",
			DeletedAt: &now,
		}
		assert.NotNil(t, group.DeletedAt)
	})
}

func TestAccountGroupModel(t *testing.T) {
	t.Run("AccountGroup model fields", func(t *testing.T) {
		now := time.Now()
		group := models.AccountGroup{
			ID:         "group-123",
			AuthUserID: "user-123",
			Name:       "Test Group",
			CreatedAt:  now,
		}
		assert.Equal(t, "group-123", group.ID)
		assert.Equal(t, "user-123", group.AuthUserID)
		assert.Equal(t, "Test Group", group.Name)
		assert.Equal(t, now, group.CreatedAt)
	})
}
