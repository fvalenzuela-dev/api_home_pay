package config

import (
	"os"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	// Save original env
	origDB := os.Getenv("DATABASE_URL")
	origClerk := os.Getenv("CLERK_SECRET_KEY")
	origWebhook := os.Getenv("CLERK_WEBHOOK_SECRET")
	origPort := os.Getenv("PORT")

	defer func() {
		os.Setenv("DATABASE_URL", origDB)
		os.Setenv("CLERK_SECRET_KEY", origClerk)
		os.Setenv("CLERK_WEBHOOK_SECRET", origWebhook)
		os.Setenv("PORT", origPort)
	}()

	// Set required env vars - use placeholders to avoid scanner warnings
	testDBURL := "localhost:5432/test"
	testClerkKey := "sk_test_placeholder"
	testWebhookSecret := "whsec_placeholder"
	testPort := "9090"

	os.Setenv("DATABASE_URL", testDBURL)
	os.Setenv("CLERK_SECRET_KEY", testClerkKey)
	os.Setenv("CLERK_WEBHOOK_SECRET", testWebhookSecret)
	os.Setenv("PORT", testPort)

	cfg, err := Load()

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DatabaseURL != testDBURL {
		t.Errorf("DatabaseURL = %v, want %v", cfg.DatabaseURL, testDBURL)
	}

	if cfg.ClerkSecretKey != testClerkKey {
		t.Errorf("ClerkSecretKey = %v, want %v", cfg.ClerkSecretKey, testClerkKey)
	}

	if cfg.ClerkWebhookSecret != testWebhookSecret {
		t.Errorf("ClerkWebhookSecret = %v, want %v", cfg.ClerkWebhookSecret, testWebhookSecret)
	}

	if cfg.Port != testPort {
		t.Errorf("Port = %v, want %v", cfg.Port, testPort)
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	// Save original env
	origDB := os.Getenv("DATABASE_URL")
	origClerk := os.Getenv("CLERK_SECRET_KEY")
	origWebhook := os.Getenv("CLERK_WEBHOOK_SECRET")

	defer func() {
		os.Setenv("DATABASE_URL", origDB)
		os.Setenv("CLERK_SECRET_KEY", origClerk)
		os.Setenv("CLERK_WEBHOOK_SECRET", origWebhook)
	}()

	// Clear required env vars
	os.Unsetenv("DATABASE_URL")
	os.Setenv("CLERK_SECRET_KEY", "sk_test_xxx")
	os.Setenv("CLERK_WEBHOOK_SECRET", "whsec_xxx")

	cfg, err := Load()

	if err == nil {
		t.Fatalf("Load() expected error for missing DATABASE_URL, got cfg = %v", cfg)
	}

	if err.Error() != "DATABASE_URL is required" {
		t.Errorf("Error = %v, want 'DATABASE_URL is required'", err)
	}
}

func TestLoad_MissingClerkSecretKey(t *testing.T) {
	// Save original env
	origDB := os.Getenv("DATABASE_URL")
	origClerk := os.Getenv("CLERK_SECRET_KEY")
	origWebhook := os.Getenv("CLERK_WEBHOOK_SECRET")

	defer func() {
		os.Setenv("DATABASE_URL", origDB)
		os.Setenv("CLERK_SECRET_KEY", origClerk)
		os.Setenv("CLERK_WEBHOOK_SECRET", origWebhook)
	}()

	os.Setenv("DATABASE_URL", "localhost:5432/test")
	os.Unsetenv("CLERK_SECRET_KEY")
	os.Setenv("CLERK_WEBHOOK_SECRET", "whsec_placeholder")

	cfg, err := Load()

	if err == nil {
		t.Fatalf("Load() expected error for missing CLERK_SECRET_KEY, got cfg = %v", cfg)
	}

	if err.Error() != "CLERK_SECRET_KEY is required" {
		t.Errorf("Error = %v, want 'CLERK_SECRET_KEY is required'", err)
	}
}

func TestLoad_MissingClerkWebhookSecret(t *testing.T) {
	// Save original env
	origDB := os.Getenv("DATABASE_URL")
	origClerk := os.Getenv("CLERK_SECRET_KEY")
	origWebhook := os.Getenv("CLERK_WEBHOOK_SECRET")

	defer func() {
		os.Setenv("DATABASE_URL", origDB)
		os.Setenv("CLERK_SECRET_KEY", origClerk)
		os.Setenv("CLERK_WEBHOOK_SECRET", origWebhook)
	}()

	os.Setenv("DATABASE_URL", "localhost:5432/test")
	os.Setenv("CLERK_SECRET_KEY", "sk_test_placeholder")
	os.Unsetenv("CLERK_WEBHOOK_SECRET")

	cfg, err := Load()

	if err == nil {
		t.Fatalf("Load() expected error for missing CLERK_WEBHOOK_SECRET, got cfg = %v", cfg)
	}

	if err.Error() != "CLERK_WEBHOOK_SECRET is required" {
		t.Errorf("Error = %v, want 'CLERK_WEBHOOK_SECRET is required'", err)
	}
}

func TestLoad_DefaultPort(t *testing.T) {
	// Save original env
	origDB := os.Getenv("DATABASE_URL")
	origClerk := os.Getenv("CLERK_SECRET_KEY")
	origWebhook := os.Getenv("CLERK_WEBHOOK_SECRET")
	origPort := os.Getenv("PORT")

	defer func() {
		os.Setenv("DATABASE_URL", origDB)
		os.Setenv("CLERK_SECRET_KEY", origClerk)
		os.Setenv("CLERK_WEBHOOK_SECRET", origWebhook)
		os.Setenv("PORT", origPort)
	}()

	os.Setenv("DATABASE_URL", "localhost:5432/test")
	os.Setenv("CLERK_SECRET_KEY", "sk_test_placeholder")
	os.Setenv("CLERK_WEBHOOK_SECRET", "whsec_placeholder")
	os.Unsetenv("PORT")

	cfg, err := Load()

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %v, want default 8080", cfg.Port)
	}
}
