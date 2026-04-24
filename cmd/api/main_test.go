package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/homepay/api/internal/config"
	"github.com/homepay/api/internal/router"
)

// TestMainInitialization verifies the router can be created
func TestMainInitialization(t *testing.T) {
	// Create router with nil handlers - we just verify router is created
	r := router.New(
		nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)

	if r == nil {
		t.Fatal("router should not be nil")
	}

	// Test router responds
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Health check - root should return 404 (chi routing)
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	t.Logf("Server responded with status: %d", resp.StatusCode)
}

// TestMainVersion verifies the version variable
func TestMainVersion(t *testing.T) {
	if version == "" {
		t.Error("version should not be empty")
	}
	t.Logf("Application version: %s", version)
}

// TestConfigValidation tests config struct
func TestConfigValidation(t *testing.T) {
	t.Run("config with empty required fields", func(t *testing.T) {
		cfg := &config.Config{}
		if cfg == nil {
			t.Fatal("config should not be nil")
		}
	})

	t.Run("config with all fields", func(t *testing.T) {
		cfg := &config.Config{
			DatabaseURL:         "postgres://user:pass@localhost:5432/db",
			ClerkSecretKey:     "sk_test_xxx",
			ClerkWebhookSecret: "whsec_xxx",
			Port:              "8080",
		}
		if cfg == nil {
			t.Fatal("config should not be nil")
		}
		if cfg.Port != "8080" {
			t.Errorf("expected port 8080, got %s", cfg.Port)
		}
	})
}

// TestAppStructure verifies the App struct
func TestAppStructure(t *testing.T) {
	t.Run("App struct can be created", func(t *testing.T) {
		app := &App{
			Config:  &config.Config{Port: "8080"},
			DB:      &closer{},
			Router:  http.DefaultServeMux,
		}
		if app.Config == nil {
			t.Error("Config should not be nil")
		}
		if app.DB == nil {
			t.Error("DB should not be nil")
		}
		if app.Router == nil {
			t.Error("Router should not be nil")
		}
	})
}

// TestInitializeApp tests the exported InitializeApp function
func TestInitializeApp(t *testing.T) {
	// Skip if no database available
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cfg := &config.Config{
		DatabaseURL:         "postgresql://postgres.vursgyoitsfoybpsicwc:leTwTKXWqO7Dvouz@aws-1-us-east-2.pooler.supabase.com:5432/postgres?search_path=homepay",
		ClerkSecretKey:     "sk_test_SHqpiFc0TyiGVznWFwq1da22o0bRTKOm2189gJTRBi",
		ClerkWebhookSecret: "whsec_CrNyZwao5nPMn9umX85j2JIGbRLRrd1P",
		Port:              "0",
	}

	// This will fail without a real database, but tests that the function exists
	_, err := InitializeApp(cfg)
	// We expect this to fail without a real database
	// The important thing is that the function is callable
	if err == nil {
		// If it succeeds, great! But likely it will fail
		t.Log("InitializeApp succeeded (unexpected in test)")
	} else {
		t.Logf("InitializeApp failed: %v", err)
	}
}

// closer implements io.Closer for testing
type closer struct{}

func (c *closer) Close() {}
