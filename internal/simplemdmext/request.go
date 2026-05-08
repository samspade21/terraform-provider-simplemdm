package simplemdmext

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	simplemdm "github.com/DavidKrau/simplemdm-go-client"
)

// DoRequest executes an HTTP request against the SimpleMDM API and returns
// the response body if the status code is in the acceptable set. On a
// non-acceptable status it returns an error that *includes the response
// body*, which is critical for debugging 4xx responses (e.g. the JSON
// {"errors":[{"title":"..."}]} envelope SimpleMDM returns).
//
// This exists because the upstream simplemdm-go-client's RequestResponseXXX
// helpers read the response body once into `body`, then on the error path
// try to read the (already-drained) body again into a separate buffer —
// which always returns an empty string. That makes any non-2xx error look
// like `got a non 201 status code: 401 - URL -` with no detail.
//
// New code in this provider should call DoRequest instead of the upstream
// client's helpers when accurate error messages matter.
func DoRequest(client *simplemdm.Client, req *http.Request, acceptable ...int) ([]byte, error) {
	req.SetBasicAuth(client.APIKey, "")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("read response body: %w", readErr)
	}

	for _, code := range acceptable {
		if resp.StatusCode == code {
			return body, nil
		}
	}

	bodyStr := strings.TrimSpace(string(body))
	if bodyStr == "" {
		return nil, fmt.Errorf("unexpected status %d from %s %s", resp.StatusCode, req.Method, req.URL)
	}
	return nil, fmt.Errorf("unexpected status %d from %s %s: %s", resp.StatusCode, req.Method, req.URL, bodyStr)
}
