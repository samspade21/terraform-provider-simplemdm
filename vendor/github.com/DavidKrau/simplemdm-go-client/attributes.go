package simplemdm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// GetAttribute - Returns a specifc attribute
func (c *Client) AttributeGet(name string) (*Attribute, error) {
	url := fmt.Sprintf("https://%s/api/v1/custom_attributes/%s", c.HostName, name)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	attribute := Attribute{}
	err = json.Unmarshal(body, &attribute)
	if err != nil {
		return nil, err
	}

	return &attribute, nil
}

// CreateAttribute - Create new attribute
func (c *Client) AttributeCreate(name string, defaultValue string) (*Attribute, error) {
	url := fmt.Sprintf("https://%s/api/v1/custom_attributes/", c.HostName)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	// adding parameter name with variable name
	q.Add("name", name)
	// checking if defaultvalue exist
	if defaultValue != "" {
		// if yes adding parameter with value
		q.Add("default_value", defaultValue)
	}
	// encoding all parameters
	req.URL.RawQuery = q.Encode()

	body, err := c.RequestResponse201(req)
	if err != nil {
		return nil, err
	}

	attribute := Attribute{}
	err = json.Unmarshal(body, &attribute)
	if err != nil {
		return nil, err
	}

	return &attribute, nil
}

// UpdateAttribute - Updates an attribute
func (c *Client) AttributeUpdate(name string, defaultValue string) error {
	url := fmt.Sprintf("https://%s/api/v1/custom_attributes/%s", c.HostName, name)
	req, err := http.NewRequest(http.MethodPatch, url, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	// checking if defaultvalue exist
	if defaultValue != "" {
		// if yes adding parameter with value
		q.Add("default_value", defaultValue)
	} else {
		q.Add("default_value", "")
	}
	// encoding all parameters
	req.URL.RawQuery = q.Encode()

	body, err := c.RequestResponse204(req)
	if err != nil {
		return err
	}

	if string(body) != "" {
		return errors.New(string(body))
	}

	return nil
}

// DeleteAttribute - Deletes an attribute
func (c *Client) AttributeDelete(name string) error {
	url := fmt.Sprintf("https://%s/api/v1/custom_attributes/%s", c.HostName, name)
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

// SetAttributeForDeviceGroupAttribute - Updates an attribute for group
func (c *Client) AttributeSetAttributeForDeviceGroup(groupID string, attribute string, defaultValue string) error {
	url := fmt.Sprintf("https://%s/api/v1/device_groups/%s/custom_attribute_values/%s", c.HostName, groupID, attribute)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("value", defaultValue)
	// encoding all parameters
	req.URL.RawQuery = q.Encode()

	body, err := c.RequestResponse200(req)
	if err != nil {
		return err
	}

	response := Attribute{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	return nil
}

// GetAttributesForDeviceGroupAttribute - Returns a specifc attribute
func (c *Client) AttributeGetAttributesForDeviceGroup(groupID string) (*AttributeArray, error) {
	url := fmt.Sprintf("https://%s/api/v1/device_groups/%s/custom_attribute_values", c.HostName, groupID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	attributes := AttributeArray{}
	err = json.Unmarshal(body, &attributes)
	if err != nil {
		return nil, err
	}

	return &attributes, nil
}

// SetAttributeForDeviceGroupAttribute - Updates an attribute for group
func (c *Client) AttributeSetAttributeForDevice(deviceID string, attribute string, value string) error {
	url := fmt.Sprintf("https://%s/api/v1/devices/%s/custom_attribute_values/%s", c.HostName, deviceID, attribute)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("value", value)
	// encoding all parameters
	req.URL.RawQuery = q.Encode()

	body, err := c.RequestResponse200(req)
	if err != nil {
		return err
	}

	response := Attribute{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	return nil
}

// GetAttributesForDeviceGroupAttribute - Returns a specifc attribute
func (c *Client) AttributeGetAttributesForDevice(deviceID string) (*AttributeArray, error) {
	url := fmt.Sprintf("https://%s/api/v1/devices/%s/custom_attribute_values", c.HostName, deviceID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	attributes := AttributeArray{}
	err = json.Unmarshal(body, &attributes)
	if err != nil {
		return nil, err
	}

	return &attributes, nil
}
