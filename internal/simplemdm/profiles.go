package simplemdm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// ProfileGet - Returns a specific profile
func (c *Client) ProfileGet(profileID string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/profiles/%s", c.HostName, profileID)
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

// Assign profile to the group
func (c *Client) ProfileAssignToGroup(profileID string, deviceGroupID string) error {
	url := fmt.Sprintf("https://%s/api/v1/profiles/%s/device_groups/%s", c.HostName, profileID, deviceGroupID)
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

// Unasiign profile from group
func (c *Client) ProfileUnAssignToGroup(profileID string, deviceGroupID string) error {
	url := fmt.Sprintf("https://%s/api/v1/profiles/%s/device_groups/%s", c.HostName, profileID, deviceGroupID)
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

// Assign profile to the device
func (c *Client) ProfileAssignToDevice(profileID string, deviceID string) error {
	url := fmt.Sprintf("https://%s/api/v1/profiles/%s/devices/%s", c.HostName, profileID, deviceID)
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

// Unasiign profile from device
func (c *Client) ProfileUnAssignToDevice(profileID string, deviceID string) error {
	url := fmt.Sprintf("https://%s/api/v1/profiles/%s/devices/%s", c.HostName, profileID, deviceID)
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

// GetAllProfiles - Returns all Profiles
func (c *Client) ProfileGetAll() (*SimpleMDMArayStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/profiles/?limit=100&starting_after=0", c.HostName)
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
			url := fmt.Sprintf("https://%s/api/v1/profiles/?limit=100&starting_after=%s", c.HostName, strconv.Itoa(profiles.Data[len(profiles.Data)-1].ID))
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
