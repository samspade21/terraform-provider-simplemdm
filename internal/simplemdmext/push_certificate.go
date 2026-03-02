package simplemdmext

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	simplemdm "github.com/DavidKrau/simplemdm-go-client"
)

// PushCertificateResponse models the push certificate API response.
type PushCertificateResponse struct {
	Data struct {
		Type       string                `json:"type"`
		Attributes PushCertificateAttrs  `json:"attributes"`
	} `json:"data"`
}

// PushCertificateAttrs contains the push certificate attributes.
type PushCertificateAttrs struct {
	AppleID   string `json:"apple_id"`
	ExpiresAt string `json:"expires_at"`
	Subject   string `json:"subject"`
}

// GetPushCertificate retrieves the current push certificate details.
func GetPushCertificate(ctx context.Context, client *simplemdm.Client) (*PushCertificateResponse, error) {
	url := fmt.Sprintf("https://%s/api/v1/push_certificate", client.HostName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := client.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	var resp PushCertificateResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
