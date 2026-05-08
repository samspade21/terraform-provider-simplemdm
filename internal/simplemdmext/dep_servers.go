package simplemdmext

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	simplemdm "github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
)

// DepServerResponse models a single DEP server response.
type DepServerResponse struct {
	Data DepServerData `json:"data"`
}

// DepServerListResponse models the list of DEP servers.
type DepServerListResponse struct {
	Data    []DepServerData `json:"data"`
	HasMore bool            `json:"has_more"`
}

// DepServerData contains DEP server attributes.
type DepServerData struct {
	Type       string         `json:"type"`
	ID         int            `json:"id"`
	Attributes DepServerAttrs `json:"attributes"`
}

// DepServerAttrs contains the attributes for a DEP server.
type DepServerAttrs struct {
	ServerName       string `json:"server_name"`
	ServerUUID       string `json:"server_uuid"`
	AdminID          int    `json:"admin_id"`
	OrganizationName string `json:"organization_name"`
	Cursor           string `json:"cursor"`
	DevicesFetchedAt string `json:"devices_fetched_at"`
	TokenExpiresAt   string `json:"token_expires_at"`
	LastSyncedAt     string `json:"last_synced_at"`
}

// DepDeviceResponse models the list of DEP devices for a server.
type DepDeviceListResponse struct {
	Data    []DepDeviceData `json:"data"`
	HasMore bool            `json:"has_more"`
}

// DepDeviceData contains DEP device attributes.
type DepDeviceData struct {
	Type       string         `json:"type"`
	ID         int            `json:"id"`
	Attributes DepDeviceAttrs `json:"attributes"`
}

// DepDeviceAttrs contains the attributes for a DEP device.
type DepDeviceAttrs struct {
	SerialNumber       string `json:"serial_number"`
	Model              string `json:"model"`
	Color              string `json:"color"`
	Description        string `json:"description"`
	OS                 string `json:"os"`
	DeviceFamily       string `json:"device_family"`
	ProfileStatus      string `json:"profile_status"`
	ProfileAssignTime  string `json:"profile_assign_time"`
	ProfilePushTime    string `json:"profile_push_time"`
	DeviceAssignedDate string `json:"device_assigned_date"`
	DeviceAssignedBy   string `json:"device_assigned_by"`
}

// GetDepServer retrieves a single DEP server by ID.
func GetDepServer(ctx context.Context, client *simplemdm.Client, serverID string) (*DepServerResponse, error) {
	url := fmt.Sprintf("https://%s/api/v1/dep_servers/%s", client.HostName, serverID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := client.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	var resp DepServerResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ListDepServers retrieves all DEP servers with pagination.
func ListDepServers(ctx context.Context, client *simplemdm.Client) ([]DepServerData, error) {
	results := make([]DepServerData, 0)
	var startingAfter string

	for {
		url := fmt.Sprintf("https://%s/api/v1/dep_servers", client.HostName)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Set("limit", "100")
		if startingAfter != "" {
			q.Set("starting_after", startingAfter)
		}
		req.URL.RawQuery = q.Encode()

		body, err := client.RequestResponse200(req)
		if err != nil {
			return nil, err
		}

		var resp DepServerListResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, err
		}

		results = append(results, resp.Data...)

		if !resp.HasMore || len(resp.Data) == 0 {
			break
		}

		startingAfter = strconv.Itoa(resp.Data[len(resp.Data)-1].ID)
	}

	return results, nil
}

// DepDeviceResponseSingle models a single DEP device response.
type DepDeviceResponseSingle struct {
	Data DepDeviceData `json:"data"`
}

// GetDepDevice retrieves a single DEP device by ID under a given DEP server.
func GetDepDevice(ctx context.Context, client *simplemdm.Client, serverID, deviceID string) (*DepDeviceResponseSingle, error) {
	url := fmt.Sprintf("https://%s/api/v1/dep_servers/%s/dep_devices/%s", client.HostName, serverID, deviceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := client.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	var resp DepDeviceResponseSingle
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// SyncDepServer triggers a sync of the given DEP server with Apple. The
// SimpleMDM API responds 202 Accepted with no body.
func SyncDepServer(ctx context.Context, client *simplemdm.Client, serverID string) error {
	url := fmt.Sprintf("https://%s/api/v1/dep_servers/%s/sync", client.HostName, serverID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	if _, err := client.RequestResponse202(req); err != nil {
		return err
	}
	return nil
}

// ListDepDevices retrieves all DEP devices for a given DEP server.
func ListDepDevices(ctx context.Context, client *simplemdm.Client, serverID string) ([]DepDeviceData, error) {
	results := make([]DepDeviceData, 0)
	var startingAfter string

	for {
		url := fmt.Sprintf("https://%s/api/v1/dep_servers/%s/dep_devices", client.HostName, serverID)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Set("limit", "100")
		if startingAfter != "" {
			q.Set("starting_after", startingAfter)
		}
		req.URL.RawQuery = q.Encode()

		body, err := client.RequestResponse200(req)
		if err != nil {
			return nil, err
		}

		var resp DepDeviceListResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, err
		}

		results = append(results, resp.Data...)

		if !resp.HasMore || len(resp.Data) == 0 {
			break
		}

		startingAfter = strconv.Itoa(resp.Data[len(resp.Data)-1].ID)
	}

	return results, nil
}
