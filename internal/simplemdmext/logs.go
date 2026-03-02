package simplemdmext

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	simplemdm "github.com/DavidKrau/simplemdm-go-client"
)

// LogListResponse models the paginated log collection response.
type LogListResponse struct {
	Data    []LogData `json:"data"`
	HasMore bool      `json:"has_more"`
}

// LogData contains a log record with attributes.
type LogData struct {
	Type       string      `json:"type"`
	ID         int         `json:"id"`
	Attributes LogAttrs    `json:"attributes"`
}

// LogAttrs contains the attributes for a log entry.
type LogAttrs struct {
	Namespace string `json:"namespace"`
	Source    string `json:"source"`
	Type      string `json:"type"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	At        string `json:"at"`
}

// ListLogs retrieves all audit logs with optional filtering, walking all pages.
func ListLogs(ctx context.Context, client *simplemdm.Client) ([]LogData, error) {
	results := make([]LogData, 0)
	var startingAfter string

	for {
		url := fmt.Sprintf("https://%s/api/v1/logs", client.HostName)
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

		var resp LogListResponse
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
