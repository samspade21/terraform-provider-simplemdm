package simplemdm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// GetDeviceGroup - Returns a specifc device group
func (c *Client) DeviceGroupGet(groupID string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/device_groups/%s", c.HostName, groupID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	deviceGroup := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &deviceGroup)
	if err != nil {
		return nil, err
	}

	return &deviceGroup, nil
}

// DeleteDeviceGroup - Returns a specifc device group
func (c *Client) DeviceGroupDelete(groupID string) error {
	url := fmt.Sprintf("https://%s/api/v1/device_groups/%s", c.HostName, groupID)
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

// CreateDeviceGroup - new device group
func (c *Client) DeviceGroupCreate(name string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/device_groups/", c.HostName)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	// adding parameter name with variable name
	q.Add("name", name)
	// checking if defaultvalue exist

	// encoding all parameters
	req.URL.RawQuery = q.Encode()

	body, err := c.RequestResponse201(req)
	if err != nil {
		return nil, err
	}

	deviceGroup := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &deviceGroup)
	if err != nil {
		return nil, err
	}

	return &deviceGroup, nil
}

// UpdateDeviceGroup - update for existing group
func (c *Client) DeviceGroupUpdate(groupID string) error {
	url := fmt.Sprintf("https://%s/api/v1/device_groups/%s", c.HostName, groupID)
	req, err := http.NewRequest(http.MethodPatch, url, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	// checking if defaultvalue exist
	// if defaultvalue != "" {
	// 	// if yes adding parameter with value
	// 	q.Add("default_value", defaultvalue)
	// } else {
	// 	q.Add("default_value", "")
	// }
	// encoding all parameters
	req.URL.RawQuery = q.Encode()

	body, err := c.RequestResponse200(req)
	if err != nil {
		return err
	}

	if string(body) != "" {
		return errors.New(string(body))
	}

	return nil
}

// AssignDeviceToDeviceGroup - Returns a specifc device group
func (c *Client) DeviceGroupAssignDevice(deviceID string, groupID string) error {
	url := fmt.Sprintf("https://%s/api/v1/device_groups/%s/devices/%s", c.HostName, groupID, deviceID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	body, err := c.RequestResponse202(req)

	if err != nil {
		return err
	}

	if string(body) != "" {
		return errors.New(string(body))
	}

	return nil
}
