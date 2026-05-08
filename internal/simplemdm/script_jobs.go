package simplemdm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) ScriptJobCreate(scriptId string, deviceIDs []string, assignmentGroupIds []string, customAttribute string, customAttributeRegex string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/script_jobs", c.HostName)

	body := &bytes.Buffer{}
	//writer := multipart.NewWriter(body)

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()

	q.Add("script_id", scriptId)

	if len(deviceIDs) > 0 {
		q.Add("device_ids", strings.Join(deviceIDs, ","))
	}
	if len(assignmentGroupIds) > 0 {
		q.Add("assignment_group_ids", strings.Join(assignmentGroupIds, ","))
	}
	if customAttribute != "" {
		q.Add("custom_attribute", customAttribute)
	}
	if customAttributeRegex != "" {
		q.Add("custom_attribute_regex", customAttributeRegex)
	}

	req.URL.RawQuery = q.Encode()

	resBody, err := c.RequestResponse201(req)
	if err != nil {
		return nil, err
	}

	scriptJob := SimplemdmDefaultStruct{}
	err = json.Unmarshal(resBody, &scriptJob)
	if err != nil {
		return nil, err
	}

	return &scriptJob, nil
}

// ScriptJobDelete - Deletes an script job
func (c *Client) ScriptCancelJob(scriptID string) error {
	url := fmt.Sprintf("https://%s/api/v1/script_jobs/%s", c.HostName, scriptID)
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

// ScriptJobGet - Returns a specifc script job
func (c *Client) ScriptJobGet(scriptJobID string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/script_jobs/%s", c.HostName, scriptJobID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	scriptJob := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &scriptJob)
	if err != nil {
		return nil, err
	}

	return &scriptJob, nil
}
