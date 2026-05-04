package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginationParams_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    PaginationParams
		expected PaginationParams
	}{
		{
			name:  "zero page defaults to 1",
			input: PaginationParams{Page: 0, PerPage: 20},
			expected: PaginationParams{Page: 1, PerPage: 20, OrderBy: "created_at", OrderDir: "desc"},
		},
		{
			name:  "negative page defaults to 1",
			input: PaginationParams{Page: -5, PerPage: 10},
			expected: PaginationParams{Page: 1, PerPage: 10, OrderBy: "created_at", OrderDir: "desc"},
		},
		{
			name:  "zero per_page defaults to 20",
			input: PaginationParams{Page: 1, PerPage: 0},
			expected: PaginationParams{Page: 1, PerPage: 20, OrderBy: "created_at", OrderDir: "desc"},
		},
		{
			name:  "negative per_page defaults to 20",
			input: PaginationParams{Page: 1, PerPage: -1},
			expected: PaginationParams{Page: 1, PerPage: 20, OrderBy: "created_at", OrderDir: "desc"},
		},
		{
			name:  "per_page over 100 defaults to 20",
			input: PaginationParams{Page: 1, PerPage: 200},
			expected: PaginationParams{Page: 1, PerPage: 20, OrderBy: "created_at", OrderDir: "desc"},
		},
		{
			name:  "empty order_by defaults to created_at",
			input: PaginationParams{Page: 1, PerPage: 20},
			expected: PaginationParams{Page: 1, PerPage: 20, OrderBy: "created_at", OrderDir: "desc"},
		},
		{
			name:  "invalid order_dir defaults to desc",
			input: PaginationParams{Page: 1, PerPage: 20, OrderDir: "random"},
			expected: PaginationParams{Page: 1, PerPage: 20, OrderBy: "created_at", OrderDir: "desc"},
		},
		{
			name:  "valid params unchanged",
			input: PaginationParams{Page: 3, PerPage: 50, OrderBy: "name", OrderDir: "asc"},
			expected: PaginationParams{Page: 3, PerPage: 50, OrderBy: "name", OrderDir: "asc"},
		},
		{
			name:  "boundary per_page=100 is valid",
			input: PaginationParams{Page: 1, PerPage: 100, OrderBy: "id", OrderDir: "desc"},
			expected: PaginationParams{Page: 1, PerPage: 100, OrderBy: "id", OrderDir: "desc"},
		},
		{
			name:  "per_page=101 exceeds limit",
			input: PaginationParams{Page: 1, PerPage: 101},
			expected: PaginationParams{Page: 1, PerPage: 20, OrderBy: "created_at", OrderDir: "desc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Normalize()
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

func TestPaginationParams_Offset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		perPage  int
		expected int
	}{
		{"page 1", 1, 20, 0},
		{"page 2", 2, 20, 20},
		{"page 3 per_page 50", 3, 50, 100},
		{"page 5 per_page 10", 5, 10, 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PaginationParams{Page: tt.page, PerPage: tt.perPage}
			assert.Equal(t, tt.expected, p.Offset())
		})
	}
}

func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		perPage  int
		expected int
	}{
		{"zero perPage", 100, 0, 0},
		{"exact division", 100, 20, 5},
		{"remainder rounds up", 101, 20, 6},
		{"zero total", 0, 20, 0},
		{"single item", 1, 20, 1},
		{"total less than perPage", 15, 20, 1},
		{"total equals perPage", 20, 20, 1},
		{"one over perPage", 21, 20, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTotalPages(tt.total, tt.perPage)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewPaginatedResponse(t *testing.T) {
	data := []string{"a", "b", "c"}
	resp := NewPaginatedResponse(data, 2, 10, 25)

	assert.Equal(t, data, resp.Data)
	assert.Equal(t, 2, resp.Pagination.Page)
	assert.Equal(t, 10, resp.Pagination.PerPage)
	assert.Equal(t, 25, resp.Pagination.Total)
	assert.Equal(t, 3, resp.Pagination.TotalPages)
}

func TestNewPaginatedResponse_EmptyData(t *testing.T) {
	resp := NewPaginatedResponse(nil, 1, 20, 0)

	assert.Nil(t, resp.Data)
	assert.Equal(t, 0, resp.Pagination.TotalPages)
}
