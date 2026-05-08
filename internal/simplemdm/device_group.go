package simplemdm

import (
	"encoding/json"
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

	_, err = c.RequestResponse204(req)
	return err
}

// CreateDeviceGroup - new device group
func (c *Client) DeviceGroupCreate(name string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/device_groups/", c.HostName)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("name", name)

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
	// if defaultvalue != "" {
	// 	// if yes adding parameter with value
	// 	q.Add("default_value", defaultvalue)
	// } else {
	// 	q.Add("default_value", "")
	// }
	req.URL.RawQuery = q.Encode()

	_, err = c.RequestResponse200(req)
	return err
}

// AssignDeviceToDeviceGroup - Returns a specifc device group
func (c *Client) DeviceGroupAssignDevice(deviceID string, groupID string) error {
	url := fmt.Sprintf("https://%s/api/v1/device_groups/%s/devices/%s", c.HostName, groupID, deviceID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	_, err = c.RequestResponse202(req)
	return err
}
