package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginationParams_Offset(t *testing.T) {
	tests := []struct {
		name      string
		page     int
		limit    int
		expected int
	}{
		{"page 1", 1, 20, 0},
		{"page 2", 2, 20, 20},
		{"page 3", 3, 20, 40},
		{"page 10", 10, 50, 450},
		{"page 1 limit 100", 1, 100, 0},
		{"page 5 limit 10", 5, 10, 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PaginationParams{Page: tt.page, Limit: tt.limit}
			assert.Equal(t, tt.expected, p.Offset())
		})
	}
}

func TestNewPaginationMeta(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		limit        int
		total        int
		expectedPage int
		expectedTotPages int
	}{
		{"basic", 1, 20, 50, 1, 3},
		{"page 2", 2, 20, 50, 2, 3},
		{"last page", 3, 20, 50, 3, 3},
		{"more pages", 1, 10, 100, 1, 10},
		{"exact division", 1, 25, 100, 1, 4},
		{"zero total", 1, 20, 0, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := NewPaginationMeta(tt.page, tt.limit, tt.total)
			assert.Equal(t, tt.expectedPage, meta.Page)
			assert.Equal(t, tt.limit, meta.Limit)
			assert.Equal(t, tt.total, meta.Total)
			assert.Equal(t, tt.expectedTotPages, meta.TotalPages)
		})
	}
}

func TestPaginationMeta_ZeroCases(t *testing.T) {
	meta := NewPaginationMeta(0, 0, 0)
	assert.Equal(t, 1, meta.Page)
	assert.Equal(t, 20, meta.Limit)
	assert.Equal(t, 0, meta.Total)
	assert.Equal(t, 0, meta.TotalPages)
}

func TestPaginationMeta_LargePage(t *testing.T) {
	// Page beyond total pages should still work
	meta := NewPaginationMeta(100, 20, 50)
	assert.Equal(t, 100, meta.Page)
	assert.Equal(t, 5, meta.TotalPages)
}
