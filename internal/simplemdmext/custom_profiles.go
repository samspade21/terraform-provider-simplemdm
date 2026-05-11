package simplemdmext

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
)

// CustomProfileExtendedAttributes mirrors the API attributes for a custom
// configuration profile while exposing the `declarative` flag that the upstream
// client struct omits.
type CustomProfileExtendedAttributes struct {
	Name                   string `json:"name"`
	UserScope              bool   `json:"user_scope"`
	AttributeSupport       bool   `json:"attribute_support"`
	EscapeAttributes       bool   `json:"escape_attributes"`
	ReinstallAfterOsUpdate bool   `json:"reinstall_after_os_update"`
	Declarative            bool   `json:"declarative"`
	ProfileIdentifier      string `json:"profile_identifier"`
	GroupCount             int    `json:"group_count"`
	DeviceCount            int    `json:"device_count"`
	ProfileSHA             string `json:"profile_sha"`
}

// CustomProfileExtendedResponse is the JSON envelope returned by the
// /custom_configuration_profiles endpoints, including the declarative flag.
type CustomProfileExtendedResponse struct {
	Data struct {
		Type       string                          `json:"type"`
		ID         int                             `json:"id"`
		Attributes CustomProfileExtendedAttributes `json:"attributes"`
	} `json:"data"`
}

// customProfileListResponse mirrors the paged LIST endpoint response.
type customProfileListResponse struct {
	Data []struct {
		Type       string                          `json:"type"`
		ID         int                             `json:"id"`
		Attributes CustomProfileExtendedAttributes `json:"attributes"`
	} `json:"data"`
	HasMore bool `json:"has_more"`
}

// CustomProfileCreatePayload represents the writable fields for create/update.
type CustomProfileCreatePayload struct {
	Name                   string
	MobileConfig           string
	UserScope              bool
	AttributeSupport       bool
	EscapeAttributes       bool
	ReinstallAfterOsUpdate bool
	Declarative            bool
}

// CreateCustomProfile creates a custom configuration profile and returns the
// full response, including the declarative flag.
func CreateCustomProfile(ctx context.Context, client *simplemdm.Client, payload CustomProfileCreatePayload) (*CustomProfileExtendedResponse, error) {
	body, contentType, err := buildCustomProfileMultipart(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s/api/v1/custom_configuration_profiles/?%s", client.HostName, customProfileQuery(payload).Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	respBody, err := client.RequestResponse201(req)
	if err != nil {
		return nil, err
	}

	return decodeCustomProfileResponse(respBody)
}

// UpdateCustomProfile updates a custom configuration profile and returns the
// full response, including the declarative flag.
func UpdateCustomProfile(ctx context.Context, client *simplemdm.Client, profileID string, payload CustomProfileCreatePayload) (*CustomProfileExtendedResponse, error) {
	body, contentType, err := buildCustomProfileMultipart(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s/api/v1/custom_configuration_profiles/%s?%s", client.HostName, profileID, customProfileQuery(payload).Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	respBody, err := client.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	return decodeCustomProfileResponse(respBody)
}

// GetCustomProfile fetches a custom configuration profile by ID, including the
// declarative flag. The SimpleMDM API does not expose a GET-by-ID endpoint for
// custom configuration profiles, so this pages through the LIST endpoint and
// returns the matching record. Returns an error containing "not found" when
// the profile is absent so callers can detect 404-equivalent results.
func GetCustomProfile(ctx context.Context, client *simplemdm.Client, profileID string) (*CustomProfileExtendedResponse, error) {
	targetID, err := strconv.Atoi(profileID)
	if err != nil {
		return nil, fmt.Errorf("invalid custom profile ID %q: %w", profileID, err)
	}

	startingAfter := 0
	for {
		url := fmt.Sprintf("https://%s/api/v1/custom_configuration_profiles/?limit=100&starting_after=%d", client.HostName, startingAfter)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		respBody, err := client.RequestResponse200(req)
		if err != nil {
			return nil, err
		}

		var page customProfileListResponse
		if err := json.Unmarshal(respBody, &page); err != nil {
			return nil, fmt.Errorf("failed to decode custom configuration profile list response: %w", err)
		}

		for _, item := range page.Data {
			if item.ID == targetID {
				resp := &CustomProfileExtendedResponse{}
				resp.Data.Type = item.Type
				resp.Data.ID = item.ID
				resp.Data.Attributes = item.Attributes
				return resp, nil
			}
		}

		if !page.HasMore || len(page.Data) == 0 {
			return nil, fmt.Errorf("custom configuration profile %s not found", profileID)
		}

		startingAfter = page.Data[len(page.Data)-1].ID
	}
}

func buildCustomProfileMultipart(payload CustomProfileCreatePayload) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("mobileconfig", payload.Name+".mobileconfig")
	if err != nil {
		return nil, "", err
	}
	if _, err := io.Copy(part, strings.NewReader(payload.MobileConfig)); err != nil {
		return nil, "", err
	}
	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}

func customProfileQuery(payload CustomProfileCreatePayload) (q queryValues) {
	q = newQueryValues()
	q.Set("name", payload.Name)
	q.SetBool("user_scope", payload.UserScope)
	q.SetBool("attribute_support", payload.AttributeSupport)
	q.SetBool("escape_attributes", payload.EscapeAttributes)
	q.SetBool("reinstall_after_os_update", payload.ReinstallAfterOsUpdate)
	q.SetBool("declarative", payload.Declarative)
	return q
}

func decodeCustomProfileResponse(body []byte) (*CustomProfileExtendedResponse, error) {
	var resp CustomProfileExtendedResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode custom configuration profile response: %w", err)
	}
	return &resp, nil
}
