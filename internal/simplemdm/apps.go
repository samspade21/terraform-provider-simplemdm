package simplemdm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
)

// AppGet - Get specific App
func (c *Client) AppGet(id string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/apps/%s", c.HostName, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	app := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &app)
	if err != nil {
		return nil, err
	}

	return &app, nil
}

// AppCreate - Create new application
func (c *Client) AppCreate(
	appStoreId string,
	bundleId string,
	name string) (*SimplemdmDefaultStruct, error) {

	url := fmt.Sprintf("https://%s/api/v1/apps", c.HostName)

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	if len(appStoreId) > 0 {
		err := writer.WriteField("app_store_id", appStoreId)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
	}

	if len(bundleId) > 0 {
		err := writer.WriteField("bundle_id", bundleId)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
	}

	if len(name) > 0 {
		err := writer.WriteField("name", name)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
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

	req.Header.Set("Content-Type", writer.FormDataContentType())
	body, err := c.RequestResponse201(req)
	if err != nil {
		return nil, err
	}

	app := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &app)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// AppDelete - Delete an application
func (c *Client) AppDelete(appId string) error {
	url := fmt.Sprintf("https://%s/api/v1/apps/%s", c.HostName, appId)
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

// AppUpdate - Updates an application (Not work for shared app)
func (c *Client) AppUpdate(appId string, name string, deployTo string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/apps/%s", c.HostName, appId)
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	if len(name) > 0 {
		err := writer.WriteField("name", name)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
	}

	if len(deployTo) > 0 {
		err := writer.WriteField("deploy_to", deployTo)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
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

	req.Header.Set("Content-Type", writer.FormDataContentType())

	body, err := c.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	app := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &app)
	if err != nil {
		return nil, err
	}

	return &app, nil
}
