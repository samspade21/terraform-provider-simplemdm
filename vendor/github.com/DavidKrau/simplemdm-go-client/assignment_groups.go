package simplemdm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// CreateAssignmentGroup - Create new addignment group
func (c *Client) AssignmentGroupCreate(name string, autoDeploy bool, groupType string, installType string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/", c.HostName)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	// adding parameter name with variable name
	q.Add("name", name)

	switch {
	case autoDeploy:
		q.Add("auto_deploy", "true")
	default:
		q.Add("auto_deploy", "false")
	}

	q.Add("type", groupType)

	if groupType == "munki" {
		switch installType {
		case "managed_updates":
			q.Add("install_type", "managed_updates")
		case "self_serve":
			q.Add("install_type", "self_serve")
		case "default_installs":
			q.Add("install_type", "default_installs")
		default:
			q.Add("install_type", "managed")
		}
	}

	//Not in API
	//auto_deploy - true
	//type standard /opt munki
	//install_type managed /opt self_serve, default_installs, managed_updates

	// encoding all parameters
	req.URL.RawQuery = q.Encode()

	body, err := c.RequestResponse201(req)
	if err != nil {
		return nil, err
	}

	assignmentGroup := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &assignmentGroup)
	if err != nil {
		return nil, err
	}

	return &assignmentGroup, nil
}

// DeleteProfile - Deletes an profile
func (c *Client) AssignmentGroupDelete(ID string) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s", c.HostName, ID)
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

// UpdateAssignmentGroup - Updates an assignment group
func (c *Client) AssignmentGroupUpdate(name string, autoDeploy bool, ID string, groupType string, installType string) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s", c.HostName, ID)
	req, err := http.NewRequest(http.MethodPatch, url, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	// adding parameter name with variable name
	q.Add("name", name)

	switch {
	case autoDeploy:
		q.Add("auto_deploy", "true")
	default:
		q.Add("auto_deploy", "false")
	}

	if groupType == "munki" {
		switch installType {
		case "managed_updates":
			q.Add("install_type", "managed_updates")
		case "self_serve":
			q.Add("install_type", "self_serve")
		case "default_installs":
			q.Add("install_type", "default_installs")
		default:
			q.Add("install_type", "managed")
		}
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

// GetAssignmentGroup - Returns a specifc assignment group
func (c *Client) AssignmentGroupGet(ID string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s", c.HostName, ID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	assignmentGroup := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &assignmentGroup)
	if err != nil {
		return nil, err
	}

	return &assignmentGroup, nil
}

// object type is app, device, group, profile, devices
// groupid is id of the assignment app
// objectid is id of the object we want to assign to the group
func (c *Client) AssignmentGroupAssignObject(groupID string, objectID string, objectType string) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s/%s/%s", c.HostName, groupID, objectType, objectID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
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

// object type is app, device, group, profile
// groupid is id of the assignment app
// objectid is id of the object we want to remove to the group
func (c *Client) AssignmentGroupUnAssignObject(groupID string, objectID string, objectType string) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s/%s/%s", c.HostName, groupID, objectType, objectID)
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

func (c *Client) AssignmentGroupPushApps(groupID string) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s/push_apps", c.HostName, groupID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	time.Sleep(30 * time.Second)
	body, err := c.RequestResponse202(req)

	if err != nil {
		return err
	}

	if string(body) != "" {
		return errors.New(string(body))
	}

	return nil
}

func (c *Client) AssignmentGroupUpdateInstalledApps(groupID string) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s/update_apps", c.HostName, groupID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	time.Sleep(30 * time.Second)
	body, err := c.RequestResponse202(req)

	if err != nil {
		return err
	}

	if string(body) != "" {
		return errors.New(string(body))
	}

	return nil
}

func (c *Client) AssignmentGroupSyncProfiles(groupID string) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s/sync_profiles", c.HostName, groupID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	time.Sleep(30 * time.Second)
	body, err := c.RequestResponse204(req)

	if err != nil {
		return err
	}

	if string(body) != "" {
		return errors.New(string(body))
	}

	return nil
}
