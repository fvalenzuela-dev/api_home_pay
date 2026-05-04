package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/homepay/api/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	testPoolCategory   *pgxpool.Pool
	testRepoCategory   CategoryRepository
	testUserIDCategory = "test-user-integration-category"
)

func setupTestDBCategory(t *testing.T) *postgres.PostgresContainer {
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

func initTestRepoCategory(t *testing.T, pgContainer *postgres.PostgresContainer) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	time.Sleep(2 * time.Second)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	if connStr == "" {
		host, err := pgContainer.Host(ctx)
		require.NoError(t, err)
		port, err := pgContainer.MappedPort(ctx, "5432")
		require.NoError(t, err)
		connStr = fmt.Sprintf("postgres://user:pass@%s:%s/homepay_test?sslmode=disable", host, port.String())
	}

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

	testPoolCategory = pool
	testRepoCategory = NewCategoryRepository(pool)

	_, err = pool.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS homepay;

		CREATE TABLE IF NOT EXISTS homepay.categories (
			id SMALLINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			auth_user_id VARCHAR(255) NOT NULL,
			icon_web VARCHAR(100),
			icon_apk VARCHAR(100),
			color_web VARCHAR(50),
			color_apk VARCHAR(50),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ,
			UNIQUE (auth_user_id, name)
		);

		CREATE INDEX IF NOT EXISTS idx_categories_user ON homepay.categories(auth_user_id);
	`)
	require.NoError(t, err, "failed to create schema")
}

func teardownTestDBCategory(t *testing.T, pgContainer *postgres.PostgresContainer) {
	t.Helper()

	if testPoolCategory != nil {
		testPoolCategory.Close()
	}
	pgContainer.Terminate(context.Background())
}

func strPtrCategory(s string) *string {
	return &s
}

func TestCategoryRepo_Create_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBCategory(t)
	initTestRepoCategory(t, pgContainer)
	defer teardownTestDBCategory(t, pgContainer)

	t.Run("creates category successfully", func(t *testing.T) {
		req := &models.CreateCategoryRequest{
			Name: "Utilities",
		}

		category, err := testRepoCategory.Create(ctx, testUserIDCategory, req)

		require.NoError(t, err)
		require.NotNil(t, category)
		assert.Equal(t, "Utilities", category.Name)
		assert.Equal(t, testUserIDCategory, category.AuthUserID)
		assert.Greater(t, category.ID, 0)
		assert.NotEmpty(t, category.CreatedAt)
		assert.NotEmpty(t, category.UpdatedAt)
	})

	t.Run("creates category with unique name per user", func(t *testing.T) {
		req := &models.CreateCategoryRequest{
			Name: "Groceries",
		}

		category, err := testRepoCategory.Create(ctx, testUserIDCategory, req)

		require.NoError(t, err)
		require.NotNil(t, category)
		assert.Equal(t, "Groceries", category.Name)
	})

	t.Run("fails with duplicate name for same user", func(t *testing.T) {
		req := &models.CreateCategoryRequest{
			Name: "Duplicate Category",
		}

		_, err := testRepoCategory.Create(ctx, testUserIDCategory, req)
		require.NoError(t, err)

		_, err = testRepoCategory.Create(ctx, testUserIDCategory, req)

		assert.Error(t, err)
		assert.Equal(t, ErrDuplicateName, err)
	})

	t.Run("allows duplicate name for different users", func(t *testing.T) {
		req := &models.CreateCategoryRequest{
			Name: "Shared Name",
		}

		_, err := testRepoCategory.Create(ctx, testUserIDCategory, req)
		require.NoError(t, err)

		category, err := testRepoCategory.Create(ctx, "other-user-id", req)

		require.NoError(t, err)
		require.NotNil(t, category)
		assert.Equal(t, "Shared Name", category.Name)
		assert.Equal(t, "other-user-id", category.AuthUserID)
	})
}

func TestCategoryRepo_GetByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBCategory(t)
	initTestRepoCategory(t, pgContainer)
	defer teardownTestDBCategory(t, pgContainer)

	var categoryID int
	err := testPoolCategory.QueryRow(ctx, `
		INSERT INTO homepay.categories (name, auth_user_id)
		VALUES ('Test Category', $1)
		RETURNING id
	`, testUserIDCategory).Scan(&categoryID)
	require.NoError(t, err)

	t.Run("gets category by id", func(t *testing.T) {
		category, err := testRepoCategory.GetByID(ctx, categoryID, testUserIDCategory)

		require.NoError(t, err)
		require.NotNil(t, category)
		assert.Equal(t, categoryID, category.ID)
		assert.Equal(t, "Test Category", category.Name)
		assert.Equal(t, testUserIDCategory, category.AuthUserID)
	})

	t.Run("returns nil for non-existent category", func(t *testing.T) {
		category, err := testRepoCategory.GetByID(ctx, 9999, testUserIDCategory)

		assert.NoError(t, err)
		assert.Nil(t, category)
	})

	t.Run("returns nil for category of different user", func(t *testing.T) {
		category, err := testRepoCategory.GetByID(ctx, categoryID, "different-user")

		assert.NoError(t, err)
		assert.Nil(t, category)
	})

	t.Run("returns nil for deleted category", func(t *testing.T) {
		_, err := testPoolCategory.Exec(ctx, `
			UPDATE homepay.categories SET deleted_at = NOW() WHERE id = $1
		`, categoryID)
		require.NoError(t, err)

		category, err := testRepoCategory.GetByID(ctx, categoryID, testUserIDCategory)

		assert.NoError(t, err)
		assert.Nil(t, category)
	})
}

func TestCategoryRepo_GetAll_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBCategory(t)
	initTestRepoCategory(t, pgContainer)
	defer teardownTestDBCategory(t, pgContainer)

	categories := []string{"Zebra", "Apple", "Banana"}
	for _, name := range categories {
		_, err := testRepoCategory.Create(ctx, testUserIDCategory, &models.CreateCategoryRequest{Name: name})
		require.NoError(t, err)
	}

	t.Run("gets all categories with pagination", func(t *testing.T) {
		pagination := models.PaginationParams{Page: 1, Limit: 2}

		result, total, err := testRepoCategory.GetAll(ctx, testUserIDCategory, pagination)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 3, total)
		assert.Equal(t, "Apple", result[0].Name)
		assert.Equal(t, "Banana", result[1].Name)
	})

	t.Run("gets all categories without pagination limits", func(t *testing.T) {
		pagination := models.PaginationParams{Page: 1, Limit: 100}

		result, total, err := testRepoCategory.GetAll(ctx, testUserIDCategory, pagination)

		require.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, 3, total)
	})

	t.Run("returns empty for user with no categories", func(t *testing.T) {
		pagination := models.PaginationParams{Page: 1, Limit: 100}

		result, total, err := testRepoCategory.GetAll(ctx, "non-existent-user", pagination)

		require.NoError(t, err)
		assert.Len(t, result, 0)
		assert.Equal(t, 0, total)
	})

	t.Run("excludes deleted categories", func(t *testing.T) {
		var catID int
		err := testPoolCategory.QueryRow(ctx, `
			INSERT INTO homepay.categories (name, auth_user_id)
			VALUES ('To Delete', $1)
			RETURNING id
		`, testUserIDCategory).Scan(&catID)
		require.NoError(t, err)

		err = testRepoCategory.Delete(ctx, catID, testUserIDCategory)
		require.NoError(t, err)

		pagination := models.PaginationParams{Page: 1, Limit: 100}
		result, total, err := testRepoCategory.GetAll(ctx, testUserIDCategory, pagination)

		require.NoError(t, err)
		assert.Equal(t, 3, total)
		for _, cat := range result {
			assert.NotEqual(t, "To Delete", cat.Name)
		}
	})
}

func TestCategoryRepo_Update_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBCategory(t)
	initTestRepoCategory(t, pgContainer)
	defer teardownTestDBCategory(t, pgContainer)

	var categoryID int
	err := testPoolCategory.QueryRow(ctx, `
		INSERT INTO homepay.categories (name, auth_user_id)
		VALUES ('Original Name', $1)
		RETURNING id
	`, testUserIDCategory).Scan(&categoryID)
	require.NoError(t, err)

	t.Run("updates category name successfully", func(t *testing.T) {
		newName := "Updated Name"

		req := &models.UpdateCategoryRequest{
			Name: &newName,
		}

		category, err := testRepoCategory.Update(ctx, categoryID, testUserIDCategory, req)

		require.NoError(t, err)
		require.NotNil(t, category)
		assert.Equal(t, newName, category.Name)
		assert.Equal(t, categoryID, category.ID)
	})

	t.Run("returns nil for non-existent category", func(t *testing.T) {
		req := &models.UpdateCategoryRequest{
			Name: strPtrCategory("New Name"),
		}

		category, err := testRepoCategory.Update(ctx, 9999, testUserIDCategory, req)

		assert.NoError(t, err)
		assert.Nil(t, category)
	})

	t.Run("returns nil for category of different user", func(t *testing.T) {
		req := &models.UpdateCategoryRequest{
			Name: strPtrCategory("Hacked Name"),
		}

		category, err := testRepoCategory.Update(ctx, categoryID, "different-user", req)

		assert.NoError(t, err)
		assert.Nil(t, category)
	})

	t.Run("fails with duplicate name", func(t *testing.T) {
		_, err := testRepoCategory.Create(ctx, testUserIDCategory, &models.CreateCategoryRequest{Name: "Existing Category"})
		require.NoError(t, err)

		req := &models.UpdateCategoryRequest{
			Name: strPtrCategory("Existing Category"),
		}

		_, err = testRepoCategory.Update(ctx, categoryID, testUserIDCategory, req)

		assert.Error(t, err)
		assert.Equal(t, ErrDuplicateName, err)
	})

	t.Run("allows updating with same name", func(t *testing.T) {
		req := &models.UpdateCategoryRequest{
			Name: strPtrCategory("Updated Name"),
		}

		category, err := testRepoCategory.Update(ctx, categoryID, testUserIDCategory, req)

		require.NoError(t, err)
		require.NotNil(t, category)
		assert.Equal(t, "Updated Name", category.Name)
	})
}

func TestCategoryRepo_Delete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBCategory(t)
	initTestRepoCategory(t, pgContainer)
	defer teardownTestDBCategory(t, pgContainer)

	var categoryID int
	err := testPoolCategory.QueryRow(ctx, `
		INSERT INTO homepay.categories (name, auth_user_id)
		VALUES ('To Delete', $1)
		RETURNING id
	`, testUserIDCategory).Scan(&categoryID)
	require.NoError(t, err)

	t.Run("soft deletes category successfully", func(t *testing.T) {
		err := testRepoCategory.Delete(ctx, categoryID, testUserIDCategory)

		require.NoError(t, err)

		category, err := testRepoCategory.GetByID(ctx, categoryID, testUserIDCategory)
		require.NoError(t, err)
		assert.Nil(t, category)
	})

	t.Run("returns error for non-existent category", func(t *testing.T) {
		err := testRepoCategory.Delete(ctx, 9999, testUserIDCategory)

		assert.Error(t, err)
	})

	t.Run("returns error for category of different user", func(t *testing.T) {
		err := testRepoCategory.Delete(ctx, categoryID, "different-user")

		assert.Error(t, err)
	})

	t.Run("returns error for already deleted category", func(t *testing.T) {
		err := testRepoCategory.Delete(ctx, categoryID, testUserIDCategory)

		assert.Error(t, err)
	})
}
