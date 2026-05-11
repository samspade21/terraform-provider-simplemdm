package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	simplemdm "github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccPreCheck ensures acceptance tests run only when TF_ACC is enabled
// and the required authentication information is present.
func testAccPreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests require TF_ACC to be set")
	}

	if os.Getenv("SIMPLEMDM_APIKEY") == "" {
		t.Skip("Acceptance tests require SIMPLEMDM_APIKEY to be set")
	}
}

// testAccRequireEnv fetches an environment variable or skips the current
// test if the variable is not defined. Reserved for the few tenant-specific
// fixtures that genuinely cannot be discovered or self-provisioned (e.g.,
// the APNs push certificate path).
func testAccRequireEnv(t *testing.T, name string) string {
	value := os.Getenv(name)
	if value == "" {
		t.Skipf("Acceptance test requires %s to be set", name)
	}

	return value
}

// testAccGetEnv fetches an environment variable and returns its value
// or an empty string if not defined. Does not skip the test.
func testAccGetEnv(t *testing.T, name string) string {
	return os.Getenv(name)
}

// getTestClient returns a SimpleMDM client configured from environment variables
// for use in test CheckDestroy functions
func getTestClient() (*simplemdm.Client, error) {
	apiKey := os.Getenv("SIMPLEMDM_APIKEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SIMPLEMDM_APIKEY environment variable not set")
	}

	host := os.Getenv("SIMPLEMDM_HOST")
	if host == "" {
		host = "a.simplemdm.com"
	}

	return simplemdm.NewClient(host, apiKey), nil
}

// testAccCheckDestroy is a helper to verify resource destruction.
//
// SimpleMDM's DELETE endpoints return 204 immediately but the underlying
// record can stay readable through GET for a few seconds — particularly
// for devices, where the destroy check often sees the resource still
// present even though the delete succeeded. Retry up to 3 times with a
// 2s gap to absorb that window.
func testAccCheckResourceDestroyed(resourceType string, checkExists func(*simplemdm.Client, string) error) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client, err := getTestClient()
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}

			var lastErr error
			for attempt := 0; attempt < 3; attempt++ {
				lastErr = checkExists(client, rs.Primary.ID)
				if lastErr != nil && isNotFoundError(lastErr) {
					break
				}
				if attempt < 2 {
					time.Sleep(2 * time.Second)
				}
			}

			if lastErr == nil {
				return fmt.Errorf("%s %s still exists after destroy", resourceType, rs.Primary.ID)
			}
			if !isNotFoundError(lastErr) {
				return fmt.Errorf("unexpected error checking %s %s: %w", resourceType, rs.Primary.ID, lastErr)
			}
		}

		return nil
	}
}

// ---------------------------------------------------------------------------
// Discovery helpers
//
// These helpers query the SimpleMDM API and return the first matching object
// so individual tests don't need to be told the ID via an env var. Each helper
// honours an env var override so CI can pin to a specific fixture if needed,
// otherwise it falls back to the API and skips cleanly when the tenant has no
// matching object.
// ---------------------------------------------------------------------------

// envOrDiscover returns the value of the named env var if set, otherwise calls
// fn() to discover a value via the API. If discovery returns an error or empty
// string, the test is skipped with a friendly message that includes both the
// env var name (for users who want to pin a value) and the discovery hint.
func envOrDiscover(t *testing.T, envName, hint string, fn func() (string, error)) string {
	if v := os.Getenv(envName); v != "" {
		return v
	}

	value, err := fn()
	if err != nil {
		t.Skipf("Acceptance test skipped: could not discover %s (%s); set %s to override. Error: %v", hint, envName, envName, err)
	}
	if value == "" {
		t.Skipf("Acceptance test skipped: %s not found in tenant (%s); set %s to override.", hint, envName, envName)
	}

	return value
}

// findFirstDeviceID returns the ID of the first fully-enrolled device in the
// tenant. Honours SIMPLEMDM_DEVICE_ID for explicit override. Skips the test if
// the tenant has no fully-enrolled devices. "awaiting_enrollment" placeholders
// (created by simplemdm_device resource tests, etc.) are filtered out because
// the API rejects commands and most other operations against them.
func findFirstDeviceID(t *testing.T) string {
	return envOrDiscover(t, "SIMPLEMDM_DEVICE_ID", "an enrolled device", func() (string, error) {
		client, err := getTestClient()
		if err != nil {
			return "", err
		}
		devices, err := simplemdmext.ListDevices(context.Background(), client, "", true, false)
		if err != nil {
			return "", err
		}
		for _, d := range devices {
			if status, ok := d.Attributes["status"].(string); ok && status == "enrolled" {
				return strconv.Itoa(d.ID), nil
			}
		}
		return "", nil
	})
}

// findFirstDeviceGroupID returns the ID of the first device group in the
// tenant. Honours SIMPLEMDM_DEVICE_GROUP_ID for explicit override.
func findFirstDeviceGroupID(t *testing.T) string {
	return envOrDiscover(t, "SIMPLEMDM_DEVICE_GROUP_ID", "a device group", func() (string, error) {
		return discoverFirstID("device_groups")
	})
}

// findFirstInstalledAppID returns the ID of the first installed-app record
// for an enrolled device. Honours SIMPLEMDM_INSTALLED_APP_ID for override.
func findFirstInstalledAppID(t *testing.T) string {
	return envOrDiscover(t, "SIMPLEMDM_INSTALLED_APP_ID", "an installed-app record", func() (string, error) {
		client, err := getTestClient()
		if err != nil {
			return "", err
		}
		devices, err := simplemdmext.ListDevices(context.Background(), client, "", true, false)
		if err != nil {
			return "", err
		}
		for _, d := range devices {
			apps, err := simplemdmext.ListDeviceInstalledApps(context.Background(), client, strconv.Itoa(d.ID))
			if err != nil {
				continue
			}
			for _, a := range apps.Data {
				if id := a.ID.String(); id != "" {
					return id, nil
				}
			}
		}
		return "", nil
	})
}

// findFirstMunkiAppID returns the ID of the first non-app-store app in the
// tenant (i.e. an enterprise or custom B2B app, which is the only kind that
// supports Munki pkginfo). Honours SIMPLEMDM_MUNKI_APP_ID for override.
func findFirstMunkiAppID(t *testing.T) string {
	return envOrDiscover(t, "SIMPLEMDM_MUNKI_APP_ID", "a custom (non-App-Store) app", func() (string, error) {
		client, err := getTestClient()
		if err != nil {
			return "", err
		}
		host := client.HostName
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/api/v1/apps?limit=100", host), nil)
		if err != nil {
			return "", err
		}
		body, err := client.RequestResponse200(req)
		if err != nil {
			return "", err
		}
		var resp struct {
			Data []struct {
				ID         int `json:"id"`
				Attributes struct {
					AppType string `json:"app_type"`
				} `json:"attributes"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return "", err
		}
		for _, a := range resp.Data {
			if a.Attributes.AppType == "enterprise" || a.Attributes.AppType == "custom b2b" {
				return strconv.Itoa(a.ID), nil
			}
		}
		return "", nil
	})
}

// discoverFirstID returns the ID of the first object at the given collection
// endpoint (e.g. "device_groups", "scripts", "enrollments"). Returns an empty
// string if the endpoint has no objects.
func discoverFirstID(endpoint string) (string, error) {
	client, err := getTestClient()
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/api/v1/%s?limit=1", client.HostName, endpoint), nil)
	if err != nil {
		return "", err
	}
	body, err := client.RequestResponse200(req)
	if err != nil {
		return "", err
	}
	var resp struct {
		Data []struct {
			ID json.RawMessage `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	if len(resp.Data) == 0 {
		return "", nil
	}
	// IDs may be int or string in the API; render either as a string.
	raw := string(resp.Data[0].ID)
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		return raw[1 : len(raw)-1], nil
	}
	return raw, nil
}
