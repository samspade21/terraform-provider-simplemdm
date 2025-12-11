package provider

import (
	"errors"
	"testing"
)

// TestIsNotFoundError tests the isNotFoundError utility function
// that classifies errors as "not found" errors for proper resource handling.
func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "error with 404",
			err:      errors.New("Error 404: Resource not available"),
			expected: true,
		},
		{
			name:     "error with 404 in middle",
			err:      errors.New("request failed with status code 404"),
			expected: true,
		},
		{
			name:     "not found error lowercase",
			err:      errors.New("resource not found"),
			expected: true,
		},
		{
			name:     "not found error uppercase - should not match (case sensitive)",
			err:      errors.New("Resource Not Found"),
			expected: false,
		},
		{
			name:     "not found error mixed case - should not match (case sensitive)",
			err:      errors.New("Item NOT FOUND in database"),
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("connection timeout"),
			expected: false,
		},
		{
			name:     "authentication error",
			err:      errors.New("unauthorized access"),
			expected: false,
		},
		{
			name:     "empty error message",
			err:      errors.New(""),
			expected: false,
		},
		{
			name:     "error with 'found' but not 'not found'",
			err:      errors.New("item found successfully"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNotFoundError(tt.err)
			if result != tt.expected {
				t.Errorf("isNotFoundError() = %v, want %v for error: %v", result, tt.expected, tt.err)
			}
		})
	}
}