package simplemdm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

func (c *Client) CustomDeclarationDelete(ID string) error {
	url := fmt.Sprintf("https://%s/api/v1/custom_declarations/%s", c.HostName, ID)
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

func (c *Client) CustomDeclarationUpdate(name string, declarationType string, declaration string, userScope bool, attributeSupport bool, escapeAttributes bool, ID string, activationPredicate string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/custom_declarations/%s", c.HostName, ID)
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	r := strings.NewReader(declaration)
	part1, errFile1 := writer.CreateFormFile("payload", name+".json")
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
	q.Add("declaration_type", declarationType)
	q.Add("activation_predicate", activationPredicate)

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

	declarationStruct := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &declarationStruct)
	if err != nil {
		return nil, err
	}

	return &declarationStruct, nil
}

func (c *Client) CustomDeclarationCreate(name string, declarationType string, declaration string, userScope bool, attributeSupport bool, escapeAttributes bool, activationPredicate string) (*SimplemdmDefaultStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/custom_declarations/", c.HostName)
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	r := strings.NewReader(declaration)
	part1, errFile1 := writer.CreateFormFile("payload", name+".json")
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

	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	// adding parameter name with variable name
	q.Add("name", name)
	q.Add("declaration_type", declarationType)
	q.Add("activation_predicate", activationPredicate)

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

	declarationStruct := SimplemdmDefaultStruct{}
	err = json.Unmarshal(body, &declarationStruct)
	if err != nil {
		return nil, err
	}

	return &declarationStruct, nil
}

func (c *Client) CustomDeclarationDownload(ID string) (*DeclarationStruct, error) {
	url := fmt.Sprintf("https://%s/api/v1/custom_declarations/%s/download", c.HostName, ID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	declaration := DeclarationStruct{}
	err = json.Unmarshal(body, &declaration)
	if err != nil {
		return nil, err
	}

	return &declaration, nil
}

func (c *Client) CustomDeclarationAssignToDevice(ID string, deviceID string) error {
	url := fmt.Sprintf("https://%s/api/v1/custom_declarations/%s/devices/%s", c.HostName, ID, deviceID)
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

func (c *Client) CustomDeclrationUnassignFromDevice(ID string, deviceID string) error {
	url := fmt.Sprintf("https://%s/api/v1/custom_declarations/%s/devices/%s", c.HostName, ID, deviceID)
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
