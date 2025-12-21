package simplemdm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

// CreateProfile - Create new profile
func (c *Client) CustomProfileCreate(name string, mobileConfig string, userScope bool, attributeSupport bool, escapeAttributes bool, reinstallAfterOsUpdate bool) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/custom_configuration_profiles/", c.HostName)
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	r := strings.NewReader(mobileConfig)
	part1, errFile1 := writer.CreateFormFile("mobileconfig", name+".mobileconfig")
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

	//MARK: 201 response
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	// adding parameter name with variable name
	q.Add("name", name)

	switch {
	case userScope:
		q.Add("user_scope", "true")
	default:
		q.Add("user_scope", "false")
	}

	switch {
	case attributeSupport:
		q.Add("attribute_support", "true")
	default:
		q.Add("attribute_support", "false")
	}

	switch {
	case escapeAttributes:
		q.Add("escape_attributes", "true")
	default:
		q.Add("escape_attributes", "false")
	}

	switch {
	case reinstallAfterOsUpdate:
		q.Add("reinstall_after_os_update", "true")
	default:
		q.Add("reinstall_after_os_update", "false")
	}

	// defaults:
	// user_scope false
	// attribute_support false
	// escape_attributes false
	// reinstall_after_os_update false

	//Not in API
	//Auto renew SCEP issued certificates
	//Enable Declarative Management

	// encoding all parameters
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Content-Type", writer.FormDataContentType())

	body, err := c.RequestResponse201(req)
	if err != nil {
		return nil, err
	}

	profile := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// DeleteProfile - Deletes an profile
func (c *Client) CustomProfileDelete(profileID string) error {
	url := fmt.Sprintf("https://%s/api/v1/custom_configuration_profiles/%s", c.HostName, profileID)
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

// UpdateProfile - Updates an profile
func (c *Client) CustomProfileUpdate(name string, mobileConfig string, userScope bool, attributeSupport bool, escapeAttributes bool, reinstallAfterOsUpdate bool, fileSHA string, ID string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/custom_configuration_profiles/%s", c.HostName, ID)
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	r := strings.NewReader(mobileConfig)
	part1, errFile1 := writer.CreateFormFile("mobileconfig", name+".mobileconfig")
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
	case userScope:
		q.Add("user_scope", "true")
	default:
		q.Add("user_scope", "false")
	}

	switch {
	case attributeSupport:
		q.Add("attribute_support", "true")
	default:
		q.Add("attribute_support", "false")
	}

	switch {
	case escapeAttributes:
		q.Add("escape_attributes", "true")
	default:
		q.Add("escape_attributes", "false")
	}

	switch {
	case reinstallAfterOsUpdate:
		q.Add("reinstall_after_os_update", "true")
	default:
		q.Add("reinstall_after_os_update", "false")
	}

	//Not in API
	//Auto renew SCEP issued certificates
	//Enable Declarative Management

	// encoding all parameters
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Content-Type", writer.FormDataContentType())

	body, err := c.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	profile := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// GetProfile - Returns a specifc profile
func (c *Client) CustomProfileGet(profileID string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/custom_configuration_profiles/%s", c.HostName, profileID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	profile := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// GetProfileSHA - Returns a specifc profile
func (c *Client) CustomProfileSHA(profileID string) (string, string, error) {
	downloadurl := fmt.Sprintf("https://%s/api/v1/custom_configuration_profiles/%s/download", c.HostName, profileID)

	req, err := http.NewRequest(http.MethodGet, downloadurl, nil)
	if err != nil {
		return "", "", err
	}

	body, sha, err := c.RequestResponse200Profile(req)
	if err != nil {
		return "", "", err
	}

	return sha, body, nil
}

// GetAllProfiles - Returns all Profiles
func (c *Client) CustomProfileGetAll() (*SimpleMDMArayStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/custom_configuration_profiles/?limit=100&starting_after=0", c.HostName)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	profiles := SimpleMDMArayStruct{}
	err = json.Unmarshal(body, &profiles)
	if err != nil {
		return nil, err
	}

	if profiles.HasMore {
		for {
			profileloop := SimpleMDMArayStruct{}
			url := fmt.Sprintf("https://%s/api/v1/custom_configuration_profiles/?limit=100&starting_after=%s", c.HostName, strconv.Itoa(profiles.Data[len(profiles.Data)-1].ID))
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				return nil, err
			}

			body, err := c.RequestResponse200(req)
			if err != nil {
				return nil, err
			}

			err = json.Unmarshal(body, &profileloop)
			if err != nil {
				return nil, err
			}

			profiles.Data = append(profiles.Data, profileloop.Data...)

			if !profileloop.HasMore {
				profiles.HasMore = false
				break
			}
		}
	}

	return &profiles, nil
}

// AssignToDeviceGroupProfile - Returns a specifc profile
func (c *Client) CustomProfileAssignToDeviceGroup(profileID string, groupID string) error {
	url := fmt.Sprintf("https://%s/api/v1/custom_configuration_profiles/%s/device_groups/%s", c.HostName, profileID, groupID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	body, err := c.RequestResponse204or409(req)
	if err != nil {
		return err
	}

	if string(body) != "" {
		return errors.New(string(body))
	}

	return nil
}

// UnassignFromDeviceGroupProfile - Returns a specifc profile
func (c *Client) CustomProfileUnassignFromDeviceGroup(profileID string, groupID string) error {
	url := fmt.Sprintf("https://%s/api/v1/custom_configuration_profiles/%s/device_groups/%s", c.HostName, profileID, groupID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	body, err := c.RequestResponse204or409(req)
	if err != nil {
		return err
	}

	if string(body) != "" {
		return errors.New(string(body))
	}

	return nil
}

func (c *Client) CustomProfileAssignToDevice(profileID string, deviceID string) error {
	url := fmt.Sprintf("https://%s/api/v1/custom_configuration_profiles/%s/devices/%s", c.HostName, profileID, deviceID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	body, err := c.RequestResponse204or409(req)
	if err != nil {
		return err
	}

	if string(body) != "" {
		return errors.New(string(body))
	}

	return nil
}

func (c *Client) CustomProfileUnAssignToDevice(profileID string, deviceID string) error {
	url := fmt.Sprintf("https://%s/api/v1/custom_configuration_profiles/%s/devices/%s", c.HostName, profileID, deviceID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	body, err := c.RequestResponse204or409(req)
	if err != nil {
		return err
	}

	if string(body) != "" {
		return errors.New(string(body))
	}

	return nil
}
