package simplemdmext

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	simplemdm "github.com/DavidKrau/simplemdm-go-client"
)

// BulkAttributeAssignment is a single device/value pair used in
// BulkSetCustomAttributeValue.
type BulkAttributeAssignment struct {
	DeviceID string `json:"device_id"`
	Value    string `json:"value"`
}

type bulkAttributePayload struct {
	Data []BulkAttributeAssignment `json:"data"`
}

// DeviceAttributeAssignment is a single attribute/value pair used in
// SetDeviceCustomAttributeValues (multi-attribute set for one device).
type DeviceAttributeAssignment struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type deviceAttributePayload struct {
	Data []DeviceAttributeAssignment `json:"data"`
}

// SetDeviceCustomAttributeValues calls
//
//	PUT /api/v1/devices/{DEVICE_ID}/custom_attribute_values
//
// with a list of {name, value} pairs, applying multiple attribute values to
// a single device in one round-trip.
func SetDeviceCustomAttributeValues(ctx context.Context, client *simplemdm.Client, deviceID string, assignments []DeviceAttributeAssignment) error {
	payload, err := json.Marshal(deviceAttributePayload{Data: assignments})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s/api/v1/devices/%s/custom_attribute_values", client.HostName, deviceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if _, err := client.RequestResponse200(req); err != nil {
		return err
	}
	return nil
}

// BulkSetCustomAttributeValue calls
//
//	PUT /api/v1/custom_attribute_values/{ATTRIBUTE_NAME}
//
// with the given list of {device_id, value} assignments. The endpoint
// responds 200 OK on success.
func BulkSetCustomAttributeValue(ctx context.Context, client *simplemdm.Client, attributeName string, assignments []BulkAttributeAssignment) error {
	payload, err := json.Marshal(bulkAttributePayload{Data: assignments})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s/api/v1/custom_attribute_values/%s", client.HostName, attributeName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if _, err := client.RequestResponse200(req); err != nil {
		return err
	}
	return nil
}
