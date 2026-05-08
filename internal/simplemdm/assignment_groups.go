package simplemdm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// CreateAssignmentGroup - Create new addignment group
func (c *Client) AssignmentGroupCreate(name string, autoDeploy bool, priority string, appTrackLocation bool) (*SimplemdmDefaultStruct, error) {
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

	switch {
	case appTrackLocation:
		q.Add("app_track_location", "true")
	default:
		q.Add("app_track_location", "false")
	}

	q.Add("priority", priority)

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
func (c *Client) AssignmentGroupUpdate(name string, autoDeploy bool, ID string, appTrackLocation bool, priority string) error {
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

	switch {
	case appTrackLocation:
		q.Add("app_track_location", "true")
	default:
		q.Add("app_track_location", "false")
	}

	q.Add("priority", priority)

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

// object type is device, group, profile, devices
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
	body, err := c.RequestResponse202or429(req)

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
	body, err := c.RequestResponse202or429(req)

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
	body, err := c.RequestResponse204or409(req)

	if err != nil {
		return err
	}

	if string(body) != "" {
		return errors.New(string(body))
	}

	return nil
}

// AttributeGetAttributesForGroup - Returns a specifc attribute
func (c *Client) AttributeGetAttributesForGroup(groupID string) (*AttributeArray, error) {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s/custom_attribute_values", c.HostName, groupID)
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

func (c *Client) AssignmentGroupAssignApp(groupID string, appID string, deploymnetType string, InstallType string) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s/apps/%s", c.HostName, groupID, appID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()

	q.Add("deployment_type", deploymnetType)
	q.Add("install_type", InstallType)

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

func (c *Client) AssignmentGroupUnAssignApp(groupID string, appID string) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s/apps/%s", c.HostName, groupID, appID)
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
