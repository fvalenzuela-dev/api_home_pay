package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// Ensure fmt is used
var _ = fmt.Sprintf

var (
	testPool   *pgxpool.Pool
	testRepo   AccountGroupRepository
	testUserID = "test-user-integration"
)

func setupTestDB(t *testing.T) *postgres.PostgresContainer {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("homepay_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
	)
	require.NoError(t, err, "failed to start postgres container")

	return pgContainer
}

func initTestRepo(t *testing.T, pgContainer *postgres.PostgresContainer) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wait for container to be fully ready
	time.Sleep(2 * time.Second)

	// Get connection string from container settings
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	// Replace database name if needed
	if connStr == "" {
		host, err := pgContainer.Host(ctx)
		require.NoError(t, err)
		port, err := pgContainer.MappedPort(ctx, "5432")
		require.NoError(t, err)
		connStr = fmt.Sprintf("postgres://***REMOVED***%s:%s/homepay_test?sslmode=disable", host, port.String())
	}

	// Retry connection
	var pool *pgxpool.Pool
	for i := 0; i < 10; i++ {
		pool, err = pgxpool.New(ctx, connStr)
		if err == nil {
			err = pool.Ping(ctx)
			if err == nil {
				break
			}
			pool.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	require.NoError(t, err, "failed to connect to database after retries")

	testPool = pool
	testRepo = NewAccountGroupRepository(pool)

	// Create schema
	_, err = pool.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS homepay;
		
		CREATE TABLE IF NOT EXISTS homepay.account_groups (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			auth_user_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			deleted_at TIMESTAMP WITH TIME ZONE
		);
		
		CREATE INDEX IF NOT EXISTS idx_account_groups_user_id ON homepay.account_groups(auth_user_id);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_account_groups_user_name ON homepay.account_groups(auth_user_id, name) WHERE deleted_at IS NULL;
	`)
	require.NoError(t, err, "failed to create schema")
}

func teardownTestDB(t *testing.T, pgContainer *postgres.PostgresContainer) {
	t.Helper()

	if testPool != nil {
		testPool.Close()
	}
	pgContainer.Terminate(context.Background())
}

// TDD: Test first - these tests define the expected behavior

func TestAccountGroupRepo_Create_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDB(t)
	initTestRepo(t, pgContainer)
	defer teardownTestDB(t, pgContainer)

	t.Run("creates account group successfully", func(t *testing.T) {
		req := &models.CreateAccountGroupRequest{
			Name: "Test Group",
		}

		group, err := testRepo.Create(ctx, testUserID, req)

		require.NoError(t, err)
		assert.NotNil(t, group)
		assert.NotEmpty(t, group.ID)
		assert.Equal(t, testUserID, group.AuthUserID)
		assert.Equal(t, "Test Group", group.Name)
		assert.NotNil(t, group.CreatedAt)
		assert.Nil(t, group.DeletedAt)
	})

	t.Run("returns error for duplicate name", func(t *testing.T) {
		req := &models.CreateAccountGroupRequest{
			Name: "Duplicate Group",
		}

		// Create first
		_, err := testRepo.Create(ctx, testUserID, req)
		require.NoError(t, err)

		// Try duplicate
		_, err = testRepo.Create(ctx, testUserID, req)

		assert.Error(t, err)
		// Error can be "name already exists" or similar duplicate error
		assert.True(t, err == ErrDuplicateName || strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "exists"))
	})
}

func TestAccountGroupRepo_GetByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDB(t)
	initTestRepo(t, pgContainer)
	defer teardownTestDB(t, pgContainer)

	// Create test data
	req := &models.CreateAccountGroupRequest{
		Name: "GetByID Test Group",
	}
	created, err := testRepo.Create(ctx, testUserID, req)
	require.NoError(t, err)

	t.Run("finds existing group", func(t *testing.T) {
		group, err := testRepo.GetByID(ctx, created.ID, testUserID)

		require.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, created.ID, group.ID)
		assert.Equal(t, "GetByID Test Group", group.Name)
	})

	t.Run("returns nil for non-existent group", func(t *testing.T) {
		// Use a valid UUID format that doesn't exist
		group, err := testRepo.GetByID(ctx, "00000000-0000-0000-0000-000000000000", testUserID)

		require.NoError(t, err)
		assert.Nil(t, group)
	})

	t.Run("returns nil for different user", func(t *testing.T) {
		group, err := testRepo.GetByID(ctx, created.ID, "other-user")

		require.NoError(t, err)
		assert.Nil(t, group)
	})
}

func TestAccountGroupRepo_GetAll_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDB(t)
	initTestRepo(t, pgContainer)
	defer teardownTestDB(t, pgContainer)

	// Create multiple groups
	groups := []string{"Alpha Group", "Beta Group", "Gamma Group"}
	for _, name := range groups {
		req := &models.CreateAccountGroupRequest{Name: name}
		_, err := testRepo.Create(ctx, testUserID, req)
		require.NoError(t, err)
	}

	// Create for different user
	_, err := testRepo.Create(ctx, "other-user", &models.CreateAccountGroupRequest{Name: "Other User Group"})
	require.NoError(t, err)

	t.Run("returns all groups for user", func(t *testing.T) {
		result, total, err := testRepo.GetAll(ctx, testUserID, models.PaginationParams{Page: 1, Limit: 10})

		require.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, 3, total)
	})

	t.Run("respects pagination", func(t *testing.T) {
		result, total, err := testRepo.GetAll(ctx, testUserID, models.PaginationParams{Page: 1, Limit: 2})

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 3, total)
	})

	t.Run("orders by name", func(t *testing.T) {
		result, _, err := testRepo.GetAll(ctx, testUserID, models.PaginationParams{Page: 1, Limit: 10})

		require.NoError(t, err)
		assert.Equal(t, "Alpha Group", result[0].Name)
		assert.Equal(t, "Beta Group", result[1].Name)
		assert.Equal(t, "Gamma Group", result[2].Name)
	})
}

func TestAccountGroupRepo_Update_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDB(t)
	initTestRepo(t, pgContainer)
	defer teardownTestDB(t, pgContainer)

	// Create test data
	created, err := testRepo.Create(ctx, testUserID, &models.CreateAccountGroupRequest{Name: "Original Name"})
	require.NoError(t, err)

	t.Run("updates group name", func(t *testing.T) {
		newName := "Updated Name"
		req := &models.UpdateAccountGroupRequest{Name: &newName}

		group, err := testRepo.Update(ctx, created.ID, testUserID, req)

		require.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, "Updated Name", group.Name)
	})

	t.Run("partial update with nil keeps existing value", func(t *testing.T) {
		req := &models.UpdateAccountGroupRequest{Name: nil}

		group, err := testRepo.Update(ctx, created.ID, testUserID, req)

		require.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, "Updated Name", group.Name) // Should keep previous value
	})

	t.Run("returns nil for non-existent", func(t *testing.T) {
		req := &models.UpdateAccountGroupRequest{}
		// Use a valid UUID format that doesn't exist
		group, err := testRepo.Update(ctx, "00000000-0000-0000-0000-000000000000", testUserID, req)

		require.NoError(t, err)
		assert.Nil(t, group)
	})
}

func TestAccountGroupRepo_SoftDelete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDB(t)
	initTestRepo(t, pgContainer)
	defer teardownTestDB(t, pgContainer)

	// Create test data
	created, err := testRepo.Create(ctx, testUserID, &models.CreateAccountGroupRequest{Name: "To Delete"})
	require.NoError(t, err)

	t.Run("soft deletes group", func(t *testing.T) {
		err := testRepo.SoftDelete(ctx, created.ID, testUserID)

		require.NoError(t, err)

		// Verify it's not found anymore
		group, err := testRepo.GetByID(ctx, created.ID, testUserID)
		require.NoError(t, err)
		assert.Nil(t, group)
	})

	t.Run("returns error for non-existent", func(t *testing.T) {
		err := testRepo.SoftDelete(ctx, "non-existent", testUserID)

		assert.Error(t, err)
	})
}
