package database

import (
	"context"
	"testing"
)

func TestConnect_InvalidURL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Test with invalid URL
	_, err := Connect(ctx, "invalid-url")

	if err == nil {
		t.Fatal("Connect() expected error for invalid URL")
	}
}

func TestConnect_InvalidDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Test with non-existent database
	_, err := Connect(ctx, "postgres://***REMOVED***:5432/nonexistent-db")

	if err == nil {
		t.Fatal("Connect() expected error for non-existent database")
	}
}

func TestConnect_FunctionExists(t *testing.T) {
	// Test that Connect function is defined
	// We can't compare a function to nil directly in Go, but we can verify it exists
	// by checking the package compiles
	_ = Connect
}
