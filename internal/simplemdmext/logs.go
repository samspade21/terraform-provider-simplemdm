package simplemdmext

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	simplemdm "github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
)

// LogListResponse models the paginated log collection response.
type LogListResponse struct {
	Data    []LogData `json:"data"`
	HasMore bool      `json:"has_more"`
}

// LogSingleResponse models the response for a single log lookup.
type LogSingleResponse struct {
	Data LogData `json:"data"`
}

// LogData contains a log record with attributes.
type LogData struct {
	Type       string   `json:"type"`
	ID         string   `json:"id"`
	Attributes LogAttrs `json:"attributes"`
}

// LogAttrs contains the attributes for a log entry. The SimpleMDM API may
// return level either as an integer (in collection responses) or as a string
// (in single-log responses); we accept both via a custom unmarshaler and
// expose the value as a string.
type LogAttrs struct {
	Namespace string          `json:"namespace"`
	Source    string          `json:"source"`
	EventType string          `json:"event_type"`
	Level     string          `json:"level"`
	Message   string          `json:"message"`
	At        string          `json:"at"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
}

// UnmarshalJSON allows level to be an int or a string in the wire format.
func (a *LogAttrs) UnmarshalJSON(b []byte) error {
	type wire struct {
		Namespace string          `json:"namespace"`
		Source    string          `json:"source"`
		EventType string          `json:"event_type"`
		Level     json.RawMessage `json:"level"`
		Message   string          `json:"message"`
		At        string          `json:"at"`
		Metadata  json.RawMessage `json:"metadata,omitempty"`
	}
	var w wire
	if err := json.Unmarshal(b, &w); err != nil {
		return err
	}
	a.Namespace = w.Namespace
	a.Source = w.Source
	a.EventType = w.EventType
	a.Message = w.Message
	a.At = w.At
	a.Metadata = w.Metadata

	if len(w.Level) == 0 {
		a.Level = ""
		return nil
	}
	// First try string.
	var s string
	if err := json.Unmarshal(w.Level, &s); err == nil {
		a.Level = s
		return nil
	}
	// Then number.
	var n json.Number
	if err := json.Unmarshal(w.Level, &n); err == nil {
		a.Level = n.String()
		return nil
	}
	a.Level = string(w.Level)
	return nil
}

// ListLogs retrieves all audit logs, walking all pages.
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

		startingAfter = resp.Data[len(resp.Data)-1].ID
	}

	return results, nil
}

// GetLog retrieves a specific log entry by ID.
func GetLog(ctx context.Context, client *simplemdm.Client, logID string) (*LogSingleResponse, error) {
	url := fmt.Sprintf("https://%s/api/v1/logs/%s", client.HostName, logID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := client.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	var resp LogSingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
