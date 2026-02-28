package simplemdmext

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	simplemdm "github.com/DavidKrau/simplemdm-go-client"
)

// AccountResponse models the payload returned from the SimpleMDM account endpoint.
type AccountResponse struct {
	Data struct {
		Type       string             `json:"type"`
		ID         int                `json:"id"`
		Attributes AccountAttributes  `json:"attributes"`
	} `json:"data"`
}

// AccountAttributes contains the account attributes returned by the API.
type AccountAttributes struct {
	Name                  string `json:"name"`
	DefaultDeviceGroupID  *int   `json:"default_device_group_id"`
	CarrierActivation     bool   `json:"carrier_activation"`
	DepEnabled            bool   `json:"dep_enabled"`
	AppUpdatesEnabled     bool   `json:"app_updates_enabled"`
	DefaultApnsCert       string `json:"default_apns_cert"`
}

// GetAccount retrieves the SimpleMDM account information.
func GetAccount(ctx context.Context, client *simplemdm.Client) (*AccountResponse, error) {
	url := fmt.Sprintf("https://%s/api/v1/account", client.HostName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := client.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	var resp AccountResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
