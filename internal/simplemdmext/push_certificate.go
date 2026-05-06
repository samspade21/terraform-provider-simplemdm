package simplemdmext

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"

	simplemdm "github.com/DavidKrau/simplemdm-go-client"
)

// PushCertificateResponse models the push certificate API response.
type PushCertificateResponse struct {
	Data struct {
		Type       string               `json:"type"`
		Attributes PushCertificateAttrs `json:"attributes"`
	} `json:"data"`
}

// PushCertificateAttrs contains the push certificate attributes.
type PushCertificateAttrs struct {
	AppleID   string `json:"apple_id"`
	ExpiresAt string `json:"expires_at"`
	Subject   string `json:"subject"`
}

// PushCertificateSCSRResponse is returned by GET /push_certificate/scsr.
type PushCertificateSCSRResponse struct {
	Data string `json:"data"`
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

// GetPushCertificateSCSR retrieves the signed CSR (a base64-encoded plist) for
// uploading to the Apple Push Certificates Portal.
func GetPushCertificateSCSR(ctx context.Context, client *simplemdm.Client) (*PushCertificateSCSRResponse, error) {
	url := fmt.Sprintf("https://%s/api/v1/push_certificate/scsr", client.HostName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := client.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	var resp PushCertificateSCSRResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// UploadPushCertificate uploads a new APNs certificate. The cert bytes are
// sent as multipart/form-data under the "file" field; appleID is optional.
func UploadPushCertificate(ctx context.Context, client *simplemdm.Client, certBytes []byte, appleID string) (*PushCertificateResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fw, err := writer.CreateFormFile("file", "push_certificate.pem")
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(certBytes); err != nil {
		return nil, err
	}
	if appleID != "" {
		if err := writer.WriteField("apple_id", appleID); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s/api/v1/push_certificate", client.HostName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	respBody, err := client.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	var resp PushCertificateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
