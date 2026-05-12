package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSetElementsToStringSliceHandlesNullAndUnknown(t *testing.T) {
	nullSet := types.SetNull(types.StringType)
	if result := setElementsToStringSlice(nullSet); len(result) != 0 {
		t.Fatalf("expected empty slice for null set, got %v", result)
	}

	unknownSet := types.SetUnknown(types.StringType)
	if result := setElementsToStringSlice(unknownSet); len(result) != 0 {
		t.Fatalf("expected empty slice for unknown set, got %v", result)
	}
}

func TestSetElementsToStringSliceFiltersInvalidElements(t *testing.T) {
	set := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("alpha"),
		types.StringUnknown(),
		types.StringNull(),
		types.StringValue("beta"),
	})

	result := setElementsToStringSlice(set)
	if len(result) != 2 {
		t.Fatalf("expected two valid values, got %d", len(result))
	}

	if result[0] != "alpha" || result[1] != "beta" {
		t.Fatalf("unexpected values returned: %v", result)
	}
}

func TestUpdateAssignmentGroupObjectsSkipsWhenNoData(t *testing.T) {
	err := updateAssignmentGroupObjects(
		context.Background(),
		nil,
		"group-id",
		types.SetUnknown(types.StringType),
		types.SetNull(types.StringType),
		"apps",
		false,
	)
	if err != nil {
		t.Fatalf("expected no error when sets are empty/unknown, got %v", err)
	}
}

func TestUpdateAssignmentGroupObjectsSkipsWhenPlanUnknown(t *testing.T) {
	state := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("123"),
	})

	err := updateAssignmentGroupObjects(
		context.Background(),
		nil,
		"group-id",
		state,
		types.SetUnknown(types.StringType),
		"apps",
		false,
	)
	if err != nil {
		t.Fatalf("expected no error when plan set is unknown, got %v", err)
	}
}

func TestSetOptionalBool(t *testing.T) {
	tests := []struct {
		name          string
		value         *bool
		expectInQuery bool
		expectedValue string
	}{
		{
			name:          "true value",
			value:         boolPtr(true),
			expectInQuery: true,
			expectedValue: "true",
		},
		{
			name:          "false value",
			value:         boolPtr(false),
			expectInQuery: true,
			expectedValue: "false",
		},
		{
			name:          "nil pointer",
			value:         nil,
			expectInQuery: false,
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := make(map[string][]string)
			setOptionalBool(values, "test_key", tt.value)

			if tt.expectInQuery {
				if val, exists := values["test_key"]; !exists {
					t.Error("expected key 'test_key' to be present in query values")
				} else if len(val) != 1 || val[0] != tt.expectedValue {
					t.Errorf("query value = %v, want [%s]", val, tt.expectedValue)
				}
			} else {
				if _, exists := values["test_key"]; exists {
					t.Error("expected key 'test_key' not to be present in query values")
				}
			}
		})
	}
}

func TestSetOptionalString(t *testing.T) {
	tests := []struct {
		name          string
		value         *string
		expectInQuery bool
		expectedValue string
	}{
		{
			name:          "non-empty string",
			value:         stringPtr("test_value"),
			expectInQuery: true,
			expectedValue: "test_value",
		},
		{
			name:          "empty string",
			value:         stringPtr(""),
			expectInQuery: false,
			expectedValue: "",
		},
		{
			name:          "nil pointer",
			value:         nil,
			expectInQuery: false,
			expectedValue: "",
		},
		{
			name:          "whitespace string",
			value:         stringPtr("   "),
			expectInQuery: true,
			expectedValue: "   ",
		},
		{
			name:          "string with special characters",
			value:         stringPtr("test&value=123"),
			expectInQuery: true,
			expectedValue: "test&value=123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := make(map[string][]string)
			setOptionalString(values, "test_key", tt.value)

			if tt.expectInQuery {
				if val, exists := values["test_key"]; !exists {
					t.Error("expected key 'test_key' to be present in query values")
				} else if len(val) != 1 || val[0] != tt.expectedValue {
					t.Errorf("query value = %v, want [%s]", val, tt.expectedValue)
				}
			} else {
				if _, exists := values["test_key"]; exists {
					t.Error("expected key 'test_key' not to be present in query values")
				}
			}
		})
	}
}
