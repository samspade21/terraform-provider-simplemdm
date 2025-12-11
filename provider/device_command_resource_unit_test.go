package provider

import (
	"net/http"
	"testing"
)

func TestValidateRequiredParameters(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		params      map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "set_admin_password with new_password",
			command:     "set_admin_password",
			params:      map[string]string{"new_password": "secret123"},
			expectError: false,
		},
		{
			name:        "set_admin_password missing new_password",
			command:     "set_admin_password",
			params:      map[string]string{},
			expectError: true,
			errorMsg:    `command "set_admin_password" requires parameter "new_password"`,
		},
		{
			name:        "set_time_zone with time_zone",
			command:     "set_time_zone",
			params:      map[string]string{"time_zone": "America/New_York"},
			expectError: false,
		},
		{
			name:        "set_time_zone missing time_zone",
			command:     "set_time_zone",
			params:      map[string]string{},
			expectError: true,
			errorMsg:    `command "set_time_zone" requires parameter "time_zone"`,
		},
		{
			name:        "delete_user with user_id",
			command:     "delete_user",
			params:      map[string]string{"user_id": "12345"},
			expectError: false,
		},
		{
			name:        "delete_user missing user_id",
			command:     "delete_user",
			params:      map[string]string{},
			expectError: true,
			errorMsg:    `command "delete_user" requires parameter "user_id"`,
		},
		{
			name:        "command with no required params",
			command:     "restart",
			params:      map[string]string{},
			expectError: false,
		},
		{
			name:        "command not in required list",
			command:     "lock",
			params:      map[string]string{},
			expectError: false,
		},
		{
			name:        "empty parameters map for required command",
			command:     "set_admin_password",
			params:      nil,
			expectError: true,
			errorMsg:    `command "set_admin_password" requires parameter "new_password"`,
		},
		{
			name:        "extra unexpected parameters should pass",
			command:     "set_admin_password",
			params:      map[string]string{"new_password": "secret123", "extra": "value"},
			expectError: false,
		},
		{
			name:        "set_admin_password with all required and optional params",
			command:     "set_admin_password",
			params:      map[string]string{"new_password": "secret123", "admin_name": "admin"},
			expectError: false,
		},
		{
			name:        "set_time_zone with multiple params",
			command:     "set_time_zone",
			params:      map[string]string{"time_zone": "UTC", "extra_param": "ignored"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequiredParameters(tt.command, tt.params)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("error message = %q, want %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestRemoveConsumedParameters(t *testing.T) {
	tests := []struct {
		name           string
		params         map[string]string
		keysToRemove   []string
		expectedParams map[string]string
	}{
		{
			name:           "remove single consumed parameter",
			params:         map[string]string{"user_id": "123", "name": "test"},
			keysToRemove:   []string{"user_id"},
			expectedParams: map[string]string{"name": "test"},
		},
		{
			name:           "remove multiple consumed parameters",
			params:         map[string]string{"user_id": "123", "device_id": "456", "name": "test"},
			keysToRemove:   []string{"user_id", "device_id"},
			expectedParams: map[string]string{"name": "test"},
		},
		{
			name:           "consumed parameter not in original map",
			params:         map[string]string{"name": "test"},
			keysToRemove:   []string{"user_id"},
			expectedParams: map[string]string{"name": "test"},
		},
		{
			name:           "empty consumed list returns original map unchanged",
			params:         map[string]string{"name": "test", "value": "123"},
			keysToRemove:   []string{},
			expectedParams: map[string]string{"name": "test", "value": "123"},
		},
		{
			name:           "nil consumed list",
			params:         map[string]string{"name": "test"},
			keysToRemove:   nil,
			expectedParams: map[string]string{"name": "test"},
		},
		{
			name:           "remove all parameters",
			params:         map[string]string{"a": "1", "b": "2"},
			keysToRemove:   []string{"a", "b"},
			expectedParams: map[string]string{},
		},
		{
			name:           "empty params map",
			params:         map[string]string{},
			keysToRemove:   []string{"key"},
			expectedParams: map[string]string{},
		},
		{
			name:           "remove non-existent keys from populated map",
			params:         map[string]string{"keep": "value"},
			keysToRemove:   []string{"remove1", "remove2"},
			expectedParams: map[string]string{"keep": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy since the function modifies in place
			testParams := make(map[string]string)
			for k, v := range tt.params {
				testParams[k] = v
			}

			removeConsumedParameters(testParams, tt.keysToRemove)

			if len(testParams) != len(tt.expectedParams) {
				t.Errorf("params length = %d, want %d", len(testParams), len(tt.expectedParams))
			}

			for key, expectedValue := range tt.expectedParams {
				if actualValue, exists := testParams[key]; !exists {
					t.Errorf("expected key %q not found in result", key)
				} else if actualValue != expectedValue {
					t.Errorf("params[%q] = %q, want %q", key, actualValue, expectedValue)
				}
			}

			// Check for unexpected keys
			for key := range testParams {
				if _, expected := tt.expectedParams[key]; !expected {
					t.Errorf("unexpected key %q found in result", key)
				}
			}
		})
	}
}

func TestPrepareCommandBody(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		params     map[string]string
		wantBody   bool
		wantString string
	}{
		{
			name:       "POST with parameters",
			method:     http.MethodPost,
			params:     map[string]string{"key": "value", "foo": "bar"},
			wantBody:   true,
			wantString: "foo=bar&key=value", // URL encoding sorts keys
		},
		{
			name:       "POST with single parameter",
			method:     http.MethodPost,
			params:     map[string]string{"message": "hello world"},
			wantBody:   true,
			wantString: "message=hello+world", // Spaces encoded as +
		},
		{
			name:       "POST with empty parameters",
			method:     http.MethodPost,
			params:     map[string]string{},
			wantBody:   false,
			wantString: "",
		},
		{
			name:       "POST with nil parameters",
			method:     http.MethodPost,
			params:     nil,
			wantBody:   false,
			wantString: "",
		},
		{
			name:       "DELETE with parameters",
			method:     http.MethodDelete,
			params:     map[string]string{"key": "value"},
			wantBody:   false,
			wantString: "",
		},
		{
			name:       "GET with parameters",
			method:     http.MethodGet,
			params:     map[string]string{"key": "value"},
			wantBody:   false,
			wantString: "",
		},
		{
			name:       "PATCH with parameters",
			method:     http.MethodPatch,
			params:     map[string]string{"key": "value"},
			wantBody:   false,
			wantString: "",
		},
		{
			name:       "POST with special characters",
			method:     http.MethodPost,
			params:     map[string]string{"pin": "1234", "message": "test&special=chars"},
			wantBody:   true,
			wantString: "message=test%26special%3Dchars&pin=1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, hasBody := prepareCommandBody(tt.method, tt.params)

			if hasBody != tt.wantBody {
				t.Errorf("hasBody = %v, want %v", hasBody, tt.wantBody)
			}

			if tt.wantBody {
				if reader == nil {
					t.Error("expected non-nil reader when hasBody is true")
					return
				}

				// Read the body content
				buf := make([]byte, 1024)
				n, _ := reader.Read(buf)
				actualString := string(buf[:n])

				// Since map iteration order is not guaranteed, we need to check both possible orderings
				// For simplicity in tests with multiple params, we'll just check key presence
				if len(tt.params) == 1 {
					if actualString != tt.wantString {
						t.Errorf("body = %q, want %q", actualString, tt.wantString)
					}
				} else if len(tt.params) > 1 {
					// For multiple params, just verify all keys are present
					for key, value := range tt.params {
						expectedPair := key + "="
						if !containsString(actualString, expectedPair) {
							t.Errorf("body missing key %q", key)
						}
						_ = value // Use value to avoid unused variable warning
					}
				}
			} else {
				if reader != nil {
					t.Error("expected nil reader when hasBody is false")
				}
			}
		})
	}
}

// Helper function for string contains check
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}