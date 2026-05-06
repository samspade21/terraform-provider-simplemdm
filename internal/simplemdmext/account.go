package simplemdmext

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	simplemdm "github.com/DavidKrau/simplemdm-go-client"
)

// AccountResponse models the payload returned from the SimpleMDM account endpoint.
type AccountResponse struct {
	Data struct {
		Type       string            `json:"type"`
		ID         int               `json:"id"`
		Attributes AccountAttributes `json:"attributes"`
	} `json:"data"`
}

// AccountAttributes contains the account attributes returned by the API.
type AccountAttributes struct {
	Name                  string `json:"name"`
	AppleStoreCountryCode string `json:"apple_store_country_code"`
	DefaultDeviceGroupID  *int   `json:"default_device_group_id"`
	CarrierActivation     bool   `json:"carrier_activation"`
	DepEnabled            bool   `json:"dep_enabled"`
	AppUpdatesEnabled     bool   `json:"app_updates_enabled"`
	DefaultApnsCert       string `json:"default_apns_cert"`
}

// GetAccount retrieves the SimpleMDM account information.
func GetAccount(ctx context.Context, client *simplemdm.Client) (*AccountResponse, error) {
	endpoint := fmt.Sprintf("https://%s/api/v1/account", client.HostName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
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

// UpdateAccount updates the SimpleMDM account name and/or apple store country
// code. Any field passed as an empty string is omitted from the request body.
func UpdateAccount(ctx context.Context, client *simplemdm.Client, name, appleStoreCountryCode string) (*AccountResponse, error) {
	form := url.Values{}
	if name != "" {
		form.Set("name", name)
	}
	if appleStoreCountryCode != "" {
		form.Set("apple_store_country_code", appleStoreCountryCode)
	}

	endpoint := fmt.Sprintf("https://%s/api/v1/account", client.HostName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
