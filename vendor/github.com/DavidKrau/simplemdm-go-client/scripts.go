package simplemdm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

func (c *Client) ScriptCreate(name string, variableSupport bool, scriptFile string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/scripts/", c.HostName)
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	r := strings.NewReader(scriptFile)
	part1, errFile1 := writer.CreateFormFile("file", name+".script")
	if errFile1 != nil {
		fmt.Println(errFile1)
		return nil, errFile1
	}
	_, errFile1 = io.Copy(part1, r)
	if errFile1 != nil {
		fmt.Println(errFile1)
		return nil, errFile1
	}
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	// adding parameter name with variable name
	q.Add("name", name)

	switch {
	case variableSupport:
		q.Add("variable_support", "1")
	default:
		q.Add("variable_support", "0")
	}

	// encoding all parameters
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Content-Type", writer.FormDataContentType())

	body, err := c.RequestResponse201(req)
	if err != nil {
		return nil, err
	}

	script := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &script)
	if err != nil {
		return nil, err
	}

	return &script, nil
}

// ScriptDelete - Deletes an script
func (c *Client) ScriptDelete(scriptID string) error {
	url := fmt.Sprintf("https://%s/api/v1/scripts/%s", c.HostName, scriptID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	body, err := c.RequestResponse204(req)

	if err != nil {
		return err
	}

	if string(body) != "" {
		return errors.New(string(body))
	}

	return nil
}

// ScriptUpdate - Updates an script
func (c *Client) ScriptUpdate(name string, variableSupport bool, scriptFile string, ID string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/scripts/%s", c.HostName, ID)
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	r := strings.NewReader(scriptFile)
	part1, errFile1 := writer.CreateFormFile("file", name+".script")
	if errFile1 != nil {
		fmt.Println(errFile1)
		return nil, errFile1
	}
	_, errFile1 = io.Copy(part1, r)
	if errFile1 != nil {
		fmt.Println(errFile1)
		return nil, errFile1
	}
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPatch, url, payload)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	// adding parameter name with variable name
	q.Add("name", name)

	switch {
	case variableSupport:
		q.Add("variable_support", "1")
	default:
		q.Add("variable_support", "0")
	}
	// encoding all parameters
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Content-Type", writer.FormDataContentType())

	body, err := c.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	script := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &script)
	if err != nil {
		return nil, err
	}

	return &script, nil
}

// ScriptGet - Returns a specifc script
func (c *Client) ScriptGet(scriptID string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/scripts/%s", c.HostName, scriptID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	script := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &script)
	if err != nil {
		return nil, err
	}

	return &script, nil
}
