package simplemdmext

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	simplemdm "github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
)

// InstalledAppResponse models a single installed app payload.
type InstalledAppResponse struct {
	Data struct {
		Type       string            `json:"type"`
		ID         int               `json:"id"`
		Attributes InstalledAppAttrs `json:"attributes"`
	} `json:"data"`
}

// InstalledAppAttrs is the attribute payload of /installed_apps/{id}.
type InstalledAppAttrs struct {
	Name         string `json:"name"`
	Identifier   string `json:"identifier"`
	Version      string `json:"version"`
	ShortVersion string `json:"short_version"`
	BundleSize   int64  `json:"bundle_size"`
	DynamicSize  *int64 `json:"dynamic_size"`
	Managed      bool   `json:"managed"`
	DiscoveredAt string `json:"discovered_at"`
	LastSeenAt   string `json:"last_seen_at"`
}

// GetInstalledApp retrieves a single installed app by ID.
func GetInstalledApp(ctx context.Context, client *simplemdm.Client, installedAppID string) (*InstalledAppResponse, error) {
	url := fmt.Sprintf("https://%s/api/v1/installed_apps/%s", client.HostName, installedAppID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := client.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	var resp InstalledAppResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// PushInstalledAppUpdate triggers POST /installed_apps/{id}/update.
func PushInstalledAppUpdate(ctx context.Context, client *simplemdm.Client, installedAppID string) error {
	url := fmt.Sprintf("https://%s/api/v1/installed_apps/%s/update", client.HostName, installedAppID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	if _, err := client.RequestResponse202(req); err != nil {
		return err
	}
	return nil
}

// RequestInstalledAppManagement triggers POST /installed_apps/{id}/request_management.
func RequestInstalledAppManagement(ctx context.Context, client *simplemdm.Client, installedAppID string) error {
	url := fmt.Sprintf("https://%s/api/v1/installed_apps/%s/request_management", client.HostName, installedAppID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	if _, err := client.RequestResponse202(req); err != nil {
		return err
	}
	return nil
}

// DeleteInstalledApp triggers DELETE /installed_apps/{id} (uninstalls the app).
func DeleteInstalledApp(ctx context.Context, client *simplemdm.Client, installedAppID string) error {
	url := fmt.Sprintf("https://%s/api/v1/installed_apps/%s", client.HostName, installedAppID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	if _, err := client.RequestResponse202(req); err != nil {
		return err
	}
	return nil
}
