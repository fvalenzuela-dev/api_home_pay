package tests

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var testDB *sql.DB

// SetupTestDatabase initializes the test database connection
func SetupTestDatabase() (*sql.DB, error) {
	if err := godotenv.Load("../.env.test"); err != nil {
		return nil, fmt.Errorf("failed to load .env.test: %w", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL not set in .env.test")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// CleanupTestDatabase closes the test database connection
func CleanupTestDatabase(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// RunMigrations runs database migrations for testing
func RunMigrations(db *sql.DB) error {
	migrationSQL := `
		-- Categories table
		CREATE TABLE IF NOT EXISTS categories (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, name)
		);

		-- Companies table
		CREATE TABLE IF NOT EXISTS companies (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			website_url VARCHAR(500),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, name)
		);

		-- Periods table
		CREATE TABLE IF NOT EXISTS periods (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			month_number INTEGER NOT NULL,
			year_number INTEGER NOT NULL,
			UNIQUE(user_id, month_number, year_number)
		);

		-- Service accounts table
		CREATE TABLE IF NOT EXISTS service_accounts (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			company_id INTEGER NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
			account_identifier VARCHAR(255) NOT NULL,
			alias VARCHAR(255),
			UNIQUE(user_id, company_id, account_identifier)
		);

		-- Expenses table
		CREATE TABLE IF NOT EXISTS expenses (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			category_id INTEGER NOT NULL REFERENCES categories(id),
			period_id INTEGER NOT NULL REFERENCES periods(id),
			account_id INTEGER REFERENCES service_accounts(id),
			description VARCHAR(500) NOT NULL,
			due_date DATE,
			current_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
			amount_paid DECIMAL(10,2) NOT NULL DEFAULT 0,
			current_installment INTEGER DEFAULT 1,
			total_installments INTEGER DEFAULT 1,
			installment_group_id VARCHAR(255),
			is_recurring BOOLEAN DEFAULT FALSE,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- Incomes table
		CREATE TABLE IF NOT EXISTS incomes (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			period_id INTEGER NOT NULL REFERENCES periods(id),
			description VARCHAR(500) NOT NULL,
			amount DECIMAL(10,2) NOT NULL DEFAULT 0,
			is_recurring BOOLEAN DEFAULT FALSE,
			received_at DATE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err := db.Exec(migrationSQL)
	return err
}

// CleanupTables truncates all tables for a clean test state
func CleanupTables(db *sql.DB) error {
	cleanupSQL := `
		TRUNCATE TABLE incomes, expenses, service_accounts, periods, companies, categories
		RESTART IDENTITY CASCADE;
	`
	_, err := db.Exec(cleanupSQL)
	return err
}

// TestMain is the entry point for integration tests
func TestMain(m *testing.M) {
	var err error
	testDB, err = SetupTestDatabase()
	if err != nil {
		fmt.Printf("Failed to setup test database: %v\n", err)
		os.Exit(1)
	}

	if err := RunMigrations(testDB); err != nil {
		fmt.Printf("Failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := CleanupTables(testDB); err != nil {
		fmt.Printf("Failed to cleanup tables: %v\n", err)
	}

	if err := CleanupTestDatabase(testDB); err != nil {
		fmt.Printf("Failed to cleanup database: %v\n", err)
	}

	os.Exit(code)
}
