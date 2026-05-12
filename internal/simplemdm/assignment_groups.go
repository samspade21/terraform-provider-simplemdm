package simplemdm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// CreateAssignmentGroup - Create new addignment group
func (c *Client) AssignmentGroupCreate(name string, autoDeploy bool, priority string, appTrackLocation bool) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/", c.HostName)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("name", name)

	q.Add("auto_deploy", strconv.FormatBool(autoDeploy))

	q.Add("app_track_location", strconv.FormatBool(appTrackLocation))

	q.Add("priority", priority)

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

	_, err = c.RequestResponse204(req)
	return err
}

// UpdateAssignmentGroup - Updates an assignment group
func (c *Client) AssignmentGroupUpdate(name string, autoDeploy bool, ID string, appTrackLocation bool, priority string) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s", c.HostName, ID)
	req, err := http.NewRequest(http.MethodPatch, url, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("name", name)

	q.Add("auto_deploy", strconv.FormatBool(autoDeploy))

	q.Add("app_track_location", strconv.FormatBool(appTrackLocation))

	q.Add("priority", priority)

	req.URL.RawQuery = q.Encode()

	_, err = c.RequestResponse204(req)
	return err
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

	_, err = c.RequestResponse204(req)
	return err
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

	_, err = c.RequestResponse204(req)
	return err
}

func (c *Client) AssignmentGroupPushApps(groupID string) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s/push_apps", c.HostName, groupID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	_, err = c.RequestResponse202or429(req)
	return err
}

func (c *Client) AssignmentGroupUpdateInstalledApps(groupID string) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s/update_apps", c.HostName, groupID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	_, err = c.RequestResponse202or429(req)
	return err
}

func (c *Client) AssignmentGroupSyncProfiles(groupID string) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s/sync_profiles", c.HostName, groupID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	_, err = c.RequestResponse204or409(req)
	return err
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

// AssignmentGroupAssignApp assigns an app to an assignment group with optional
// deployment_type and install_type. Empty strings mean "use SimpleMDM
// defaults" — they are omitted from the query rather than sent as
// `deployment_type=` (which the API rejects with 400 "Deployment type can't
// be blank").
func (c *Client) AssignmentGroupAssignApp(groupID string, appID string, deploymentType string, installType string) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s/apps/%s", c.HostName, groupID, appID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	if deploymentType != "" {
		q.Add("deployment_type", deploymentType)
	}
	if installType != "" {
		q.Add("install_type", installType)
	}
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

	_, err = c.RequestResponse204(req)
	return err
}
