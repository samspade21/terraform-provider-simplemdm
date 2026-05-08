// Package simplemdm is the SimpleMDM API client used by this provider.
//
// History: this package was originally a separate module
// (github.com/DavidKrau/simplemdm-go-client) that the provider depended on
// externally. It has been vendored into the provider so we can fix bugs in
// place (notably the response-body-drain issue that caused every 4xx error
// to surface with an empty body) and add coverage for endpoints the
// upstream module didn't expose.
package simplemdm

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client holds all the information required to talk to the SimpleMDM API.
type Client struct {
	HostName   string
	APIKey     string
	httpClient *http.Client
}

// NewClient constructs a Client for the given host (e.g. "a.simplemdm.com")
// and API key.
func NewClient(hostname string, apikey string) *Client {
	return &Client{
		HostName: hostname,
		APIKey:   apikey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// do executes a request and returns the response body. If the status code is
// not in `acceptable`, the returned error includes the response body so
// callers can see the SimpleMDM JSON error envelope (e.g. {"errors":[…]}).
func (c *Client) do(req *http.Request, acceptable ...int) ([]byte, error) {
	req.SetBasicAuth(c.APIKey, "")

	// SimpleMDM has fairly aggressive rate limits; pause briefly before
	// every request to keep within them. The upstream module did the same.
	time.Sleep(1 * time.Second)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	for _, code := range acceptable {
		if res.StatusCode == code {
			return body, nil
		}
	}

	return nil, statusError(res.StatusCode, req, body, acceptable)
}

// statusError formats a non-acceptable HTTP status as a Go error, including
// the response body when present.
func statusError(statusCode int, req *http.Request, body []byte, acceptable []int) error {
	expected := "<unknown>"
	switch len(acceptable) {
	case 0:
		// nothing
	case 1:
		expected = fmt.Sprintf("%d", acceptable[0])
	default:
		expected = fmt.Sprintf("one of %v", acceptable)
	}

	if len(body) == 0 {
		return fmt.Errorf("unexpected status %d (expected %s) from %s %s", statusCode, expected, req.Method, req.URL)
	}
	return fmt.Errorf("unexpected status %d (expected %s) from %s %s: %s", statusCode, expected, req.Method, req.URL, string(body))
}

// RequestResponse200 expects the API to respond with HTTP 200 OK. Returns the
// response body or a descriptive error including the response body.
func (c *Client) RequestResponse200(req *http.Request) ([]byte, error) {
	return c.do(req, http.StatusOK)
}

// RequestResponse200Profile is a specialised helper used by the custom
// configuration profile download endpoint, which returns the raw mobileconfig
// body and a SHA-256 checksum in the ETag header.
//
// Returns (body, sha, error). The SHA comes from the ETag header — SimpleMDM
// formats it as `W/"<32-char-hex>"`, so we slice [3:35].
func (c *Client) RequestResponse200Profile(req *http.Request) (string, string, error) {
	req.SetBasicAuth(c.APIKey, "")
	time.Sleep(1 * time.Second)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", fmt.Errorf("read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return "", "", statusError(res.StatusCode, req, body, []int{http.StatusOK})
	}

	var sha string
	if etag := res.Header.Get("etag"); len(etag) >= 35 {
		sha = etag[3:35]
	}
	return string(body), sha, nil
}

// RequestResponse201 expects HTTP 201 Created.
func (c *Client) RequestResponse201(req *http.Request) ([]byte, error) {
	return c.do(req, http.StatusCreated)
}

// RequestResponse202 expects HTTP 202 Accepted.
func (c *Client) RequestResponse202(req *http.Request) ([]byte, error) {
	return c.do(req, http.StatusAccepted)
}

// RequestResponse204 expects HTTP 204 No Content.
func (c *Client) RequestResponse204(req *http.Request) ([]byte, error) {
	return c.do(req, http.StatusNoContent)
}

// RequestResponse204or409 accepts either 204 No Content (success) or 409
// Conflict (idempotent assign / unassign — already in the desired state).
// Treats both as success and returns nil.
func (c *Client) RequestResponse204or409(req *http.Request) ([]byte, error) {
	res, body, err := c.execute(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNoContent || res.StatusCode == http.StatusConflict {
		return body, nil
	}
	return nil, statusError(res.StatusCode, req, body, []int{http.StatusNoContent, http.StatusConflict})
}

// RequestResponse202or429 retries once after sleeping 30s when SimpleMDM
// returns 429 Too Many Requests, then expects 202 Accepted.
func (c *Client) RequestResponse202or429(req *http.Request) ([]byte, error) {
	res, body, err := c.execute(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusTooManyRequests {
		time.Sleep(30 * time.Second)
		res, body, err = c.execute(req)
		if err != nil {
			return nil, err
		}
	}

	if res.StatusCode == http.StatusAccepted {
		return body, nil
	}
	return nil, statusError(res.StatusCode, req, body, []int{http.StatusAccepted})
}

// execute is a low-level helper used by the variants that need to inspect
// the status code before deciding what to do (e.g. retry on 429, accept 409
// as success). Returns the response, the read-and-closed body, and any
// transport-level error.
func (c *Client) execute(req *http.Request) (*http.Response, []byte, error) {
	req.SetBasicAuth(c.APIKey, "")
	time.Sleep(1 * time.Second)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response body: %w", err)
	}
	return res, body, nil
}
