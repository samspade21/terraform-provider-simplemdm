package simplemdmext

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"

	simplemdm "github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
)

// AppInstallItem represents a single per-device install record returned from
// /apps/{APP_ID}/installs.
type AppInstallItem struct {
	Type          string         `json:"type"`
	ID            jsonNumber     `json:"id"`
	Attributes    map[string]any `json:"attributes"`
	Relationships struct {
		Device struct {
			Data struct {
				Type string     `json:"type"`
				ID   jsonNumber `json:"id"`
			} `json:"data"`
		} `json:"device"`
	} `json:"relationships"`
}

type appInstallsResponse struct {
	Data    []AppInstallItem `json:"data"`
	HasMore bool             `json:"has_more"`
}

// UploadMunkiPkgInfo replaces the Munki pkginfo XML/PLIST blob attached to an
// app via POST /apps/{APP_ID}/munki_pkginfo (multipart/form-data, field "file").
// API responds 202 Accepted.
func UploadMunkiPkgInfo(ctx context.Context, client *simplemdm.Client, appID string, fileBytes []byte, filename string) error {
	if filename == "" {
		filename = "munki_pkginfo.plist"
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fw, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	if _, err := fw.Write(fileBytes); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s/api/v1/apps/%s/munki_pkginfo", client.HostName, appID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if _, err := client.RequestResponse202(req); err != nil {
		return err
	}
	return nil
}

// DeleteMunkiPkgInfo clears the pkginfo associated with an app via
// DELETE /apps/{APP_ID}/munki_pkginfo (204 No Content).
func DeleteMunkiPkgInfo(ctx context.Context, client *simplemdm.Client, appID string) error {
	url := fmt.Sprintf("https://%s/api/v1/apps/%s/munki_pkginfo", client.HostName, appID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	if _, err := client.RequestResponse204(req); err != nil {
		return err
	}
	return nil
}

// ListAppInstalls walks paginated installs of a given app across the fleet.
func ListAppInstalls(ctx context.Context, client *simplemdm.Client, appID string) ([]AppInstallItem, error) {
	out := make([]AppInstallItem, 0)
	var startingAfter string

	for {
		url := fmt.Sprintf("https://%s/api/v1/apps/%s/installs", client.HostName, appID)
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

		var resp appInstallsResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, err
		}

		out = append(out, resp.Data...)

		if !resp.HasMore || len(resp.Data) == 0 {
			break
		}
		startingAfter = resp.Data[len(resp.Data)-1].ID.String()
	}

	return out, nil
}
