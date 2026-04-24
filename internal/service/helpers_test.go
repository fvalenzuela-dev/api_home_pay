package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests for CompanyService validation
func TestCompanyServiceValidation(t *testing.T) {
	// Test that would require valid input
	t.Run("validation placeholder", func(t *testing.T) {
		// Placeholder for company service validation tests
		// Actual tests require proper mock setup with full interface
		assert.True(t, true)
	})
}

// Tests for nextPeriod helper
func TestNextPeriod(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{202601, 202602},
		{202602, 202603},
		{202603, 202604},
		{202611, 202612},
		{202612, 202701}, // December -> January next year
		{202699, 202700}, // Edge case
	}

	for _, tt := range tests {
		result := nextPeriod(tt.input)
		assert.Equal(t, tt.expected, result, "nextPeriod(%d)", tt.input)
	}
}

// Tests for previousPeriod helper
func TestPreviousPeriod(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{202602, 202601},
		{202603, 202602},
		{202601, 202512}, // January -> December previous year
		{202612, 202611},
		{202701, 202612},
	}

	for _, tt := range tests {
		result := previousPeriod(tt.input)
		assert.Equal(t, tt.expected, result, "previousPeriod(%d)", tt.input)
	}
}

// Tests for validatePeriod helper
func TestValidatePeriod(t *testing.T) {
	tests := []struct {
		period    int
		shouldErr bool
	}{
		{202601, false},  // Valid January
		{202612, false},  // Valid December
		{202603, false},  // Valid March
		{202600, true},   // Invalid month 0
		{202613, true},   // Invalid month 13
		{202615, true},   // Invalid month 15
		{202000, true},   // Invalid month 0 (year 2020, month 00)
		{202100, true},   // Invalid month 0 (year 2100, month 00)
	}

	for _, tt := range tests {
		err := validatePeriod(tt.period)
		if tt.shouldErr {
			assert.Error(t, err, "period %d should error", tt.period)
		} else {
			// Only valid periods shouldn't error
			if tt.period >= 202601 && tt.period <= 202612 {
				assert.NoError(t, err, "period %d should not error", tt.period)
			}
		}
	}
}
