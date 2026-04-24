package repository

import (
	"context"
	"database/sql"
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
	testPoolUser   *pgxpool.Pool
	testRepoUser   UserRepository
	testUserIDUser = "test-user-integration-user"
)

func setupTestDBUser(t *testing.T) *postgres.PostgresContainer {
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

func initTestRepoUser(t *testing.T, pgContainer *postgres.PostgresContainer) {
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
		connStr = fmt.Sprintf("postgres://***REMOVED***%s:%s/homepay_test?sslmode=disable", host, port.String())
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

	testPoolUser = pool
	testRepoUser = NewUserRepository(pool)

	_, err = pool.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS homepay;

		CREATE TABLE IF NOT EXISTS homepay.users (
			auth_user_id VARCHAR(255) PRIMARY KEY,
			email VARCHAR(255),
			full_name VARCHAR(255),
			timezone VARCHAR(50) DEFAULT 'UTC',
			currency VARCHAR(10) DEFAULT 'USD',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		);

		CREATE INDEX IF NOT EXISTS idx_users_auth_user_id ON homepay.users(auth_user_id);
	`)
	require.NoError(t, err, "failed to create schema")
}

func teardownTestDBUser(t *testing.T, pgContainer *postgres.PostgresContainer) {
	t.Helper()

	if testPoolUser != nil {
		testPoolUser.Close()
	}
	pgContainer.Terminate(context.Background())
}

func TestUserRepo_Upsert_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBUser(t)
	initTestRepoUser(t, pgContainer)
	defer teardownTestDBUser(t, pgContainer)

	t.Run("inserts new user successfully", func(t *testing.T) {
		user := &models.User{
			AuthUserID: testUserIDUser,
			Email:      "test@example.com",
			FullName:   "Test User",
		}

		err := testRepoUser.Upsert(ctx, user)

		require.NoError(t, err)

		var count int
		err = testPoolUser.QueryRow(ctx, `SELECT COUNT(*) FROM homepay.users WHERE auth_user_id = $1`, testUserIDUser).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("updates existing user on conflict", func(t *testing.T) {
		initialUser := &models.User{
			AuthUserID: testUserIDUser,
			Email:      "initial@example.com",
			FullName:   "Initial Name",
		}
		err := testRepoUser.Upsert(ctx, initialUser)
		require.NoError(t, err)

		updatedUser := &models.User{
			AuthUserID: testUserIDUser,
			Email:      "updated@example.com",
			FullName:   "Updated Name",
		}

		err = testRepoUser.Upsert(ctx, updatedUser)

		require.NoError(t, err)

		var email, fullName string
		err = testPoolUser.QueryRow(ctx, `SELECT email, full_name FROM homepay.users WHERE auth_user_id = $1`, testUserIDUser).Scan(&email, &fullName)
		require.NoError(t, err)
		assert.Equal(t, "updated@example.com", email)
		assert.Equal(t, "Updated Name", fullName)
	})

	t.Run("handles empty email and full_name", func(t *testing.T) {
		user := &models.User{
			AuthUserID: "user-empty-fields",
			Email:      "",
			FullName:   "",
		}

		err := testRepoUser.Upsert(ctx, user)

		require.NoError(t, err)

		var email, fullName sql.NullString
		err = testPoolUser.QueryRow(ctx, `SELECT email, full_name FROM homepay.users WHERE auth_user_id = $1`, "user-empty-fields").Scan(&email, &fullName)
		require.NoError(t, err)
		assert.Equal(t, "", email.String)
		assert.Equal(t, "", fullName.String)
	})

	t.Run("does not update deleted user on conflict", func(t *testing.T) {
		deletedUserID := "user-deleted-conflict"

		_, err := testPoolUser.Exec(ctx, `
			INSERT INTO homepay.users (auth_user_id, email, full_name, deleted_at)
			VALUES ($1, 'deleted@example.com', 'Deleted User', NOW())
		`, deletedUserID)
		require.NoError(t, err)

		user := &models.User{
			AuthUserID: deletedUserID,
			Email:      "new@example.com",
			FullName:   "New User",
		}

		err = testRepoUser.Upsert(ctx, user)

		require.NoError(t, err)

		var email, fullName string
		var deletedAt *time.Time
		err = testPoolUser.QueryRow(ctx, `SELECT email, full_name, deleted_at FROM homepay.users WHERE auth_user_id = $1`, deletedUserID).Scan(&email, &fullName, &deletedAt)
		require.NoError(t, err)
		assert.Equal(t, "deleted@example.com", email)
		assert.Equal(t, "Deleted User", fullName)
		assert.NotNil(t, deletedAt)
	})
}

func TestUserRepo_SoftDelete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer := setupTestDBUser(t)
	initTestRepoUser(t, pgContainer)
	defer teardownTestDBUser(t, pgContainer)

	t.Run("soft deletes user successfully", func(t *testing.T) {
		userToDelete := &models.User{
			AuthUserID: "user-to-delete",
			Email:      "delete@example.com",
			FullName:   "Delete User",
		}

		err := testRepoUser.Upsert(ctx, userToDelete)
		require.NoError(t, err)

		err = testRepoUser.SoftDelete(ctx, "user-to-delete")

		require.NoError(t, err)

		var deletedAt *time.Time
		err = testPoolUser.QueryRow(ctx, `SELECT deleted_at FROM homepay.users WHERE auth_user_id = $1`, "user-to-delete").Scan(&deletedAt)
		require.NoError(t, err)
		assert.NotNil(t, deletedAt)
	})

	t.Run("does not affect non-existent user", func(t *testing.T) {
		err := testRepoUser.SoftDelete(ctx, "non-existent-user")

		require.NoError(t, err)
	})

	t.Run("does not delete already deleted user", func(t *testing.T) {
		alreadyDeletedID := "user-already-deleted"

		_, err := testPoolUser.Exec(ctx, `
			INSERT INTO homepay.users (auth_user_id, email, deleted_at)
			VALUES ($1, 'test@example.com', NOW())
		`, alreadyDeletedID)
		require.NoError(t, err)

		err = testRepoUser.SoftDelete(ctx, alreadyDeletedID)

		require.NoError(t, err)

		var count int
		err = testPoolUser.QueryRow(ctx, `SELECT COUNT(*) FROM homepay.users WHERE auth_user_id = $1 AND deleted_at IS NOT NULL`, alreadyDeletedID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}
