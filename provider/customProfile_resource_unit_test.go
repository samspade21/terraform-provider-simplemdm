package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestStringValueOrNull tests the stringValueOrNull utility function
// that converts empty strings to Terraform null values for proper state management.
func TestStringValueOrNull(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected types.String
	}{
		{
			name:     "non-empty string",
			input:    "test-value",
			expected: types.StringValue("test-value"),
		},
		{
			name:     "empty string returns null",
			input:    "",
			expected: types.StringNull(),
		},
		{
			name:     "string with spaces",
			input:    "value with spaces",
			expected: types.StringValue("value with spaces"),
		},
		{
			name:     "string with special characters",
			input:    "test@example.com",
			expected: types.StringValue("test@example.com"),
		},
		{
			name:     "numeric string",
			input:    "12345",
			expected: types.StringValue("12345"),
		},
		{
			name:     "string with newline",
			input:    "line1\nline2",
			expected: types.StringValue("line1\nline2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringValueOrNull(tt.input)

			// Compare null states
			if result.IsNull() != tt.expected.IsNull() {
				t.Errorf("stringValueOrNull() null state = %v, want %v", result.IsNull(), tt.expected.IsNull())
			}

			// If not null, compare values
			if !result.IsNull() && !tt.expected.IsNull() {
				if result.ValueString() != tt.expected.ValueString() {
					t.Errorf("stringValueOrNull() = %v, want %v", result.ValueString(), tt.expected.ValueString())
				}
			}
		})
	}
}
