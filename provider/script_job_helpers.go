package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	simplemdm "github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// scriptJobErrorHint augments a script_job create error with a tenant-aware
// diagnostic when the underlying failure is a 422 / 400 from SimpleMDM —
// those statuses indicate the targets are wrong (no devices, non-enrolled
// IDs, non-macOS IDs), which the API surfaces only as a generic message.
// For other error classes (5xx, timeouts, auth) the hint isn't relevant
// and we return "" without doing the extra ListDevices call.
//
// The check is best-effort: any error fetching the device list silently
// returns "". Script jobs run on macOS only, so non-macOS targets are
// called out specifically.
func scriptJobErrorHint(ctx context.Context, client *simplemdm.Client, originalErr error, deviceIDs []string) string {
	if originalErr == nil {
		return ""
	}
	errStr := originalErr.Error()
	if !strings.Contains(errStr, "422") && !strings.Contains(errStr, "400") {
		return ""
	}

	devices, err := simplemdmext.ListDevices(ctx, client, "", true, false)
	if err != nil {
		return ""
	}

	if len(devices) == 0 {
		return "Hint: the tenant has no devices at all. Script jobs require at least one enrolled macOS device target. Enroll a device through the SimpleMDM web UI before re-running."
	}

	enrolled := map[string]map[string]any{}
	for i := range devices {
		if status, _ := devices[i].Attributes["status"].(string); status == "enrolled" {
			enrolled[strconv.Itoa(devices[i].ID)] = devices[i].Attributes
		}
	}

	if len(enrolled) == 0 {
		return "Hint: the tenant has devices but none are in `enrolled` state (script jobs can only target enrolled devices). Wait for enrollment to complete or enroll a new device."
	}

	if len(deviceIDs) == 0 {
		return ""
	}
	var bad, nonMac []string
	for _, id := range deviceIDs {
		attrs, ok := enrolled[id]
		if !ok {
			bad = append(bad, id)
			continue
		}
		// SimpleMDM's `product_name` for macOS devices contains "Mac" (e.g.
		// "MacBookPro14,3", "iMac20,2", "Macmini8,1"). A single Contains
		// check covers MacBook / iMac / Macmini / MacPro / MacStudio.
		if pname, _ := attrs["product_name"].(string); !strings.Contains(pname, "Mac") {
			nonMac = append(nonMac, id)
		}
	}
	switch {
	case len(bad) > 0 && len(nonMac) > 0:
		return fmt.Sprintf("Hint: device IDs %s are not enrolled, and %s are non-macOS (script jobs only run on macOS).", strings.Join(bad, ", "), strings.Join(nonMac, ", "))
	case len(bad) > 0:
		return fmt.Sprintf("Hint: device IDs %s are not in `enrolled` state. Script jobs can only target enrolled devices.", strings.Join(bad, ", "))
	case len(nonMac) > 0:
		return fmt.Sprintf("Hint: device IDs %s are non-macOS. Script jobs only run on macOS devices.", strings.Join(nonMac, ", "))
	}
	return ""
}

type scriptJobDeviceDetail struct {
	ID         string
	Status     string
	StatusCode *string
	Response   *string
}

type scriptJobDetailsData struct {
	ID                   string
	ScriptName           string
	JobName              string
	JobIdentifier        string
	Status               string
	PendingCount         int64
	SuccessCount         int64
	ErroredCount         int64
	Content              string
	VariableSupport      bool
	CreatedBy            string
	CreatedAt            string
	UpdatedAt            string
	CustomAttribute      string
	CustomAttributeRegex string
	Devices              []scriptJobDeviceDetail
}

var scriptJobDeviceAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"status":      types.StringType,
	"status_code": types.StringType,
	"response":    types.StringType,
}

type scriptJobDetailsResponse struct {
	Data struct {
		ID         int `json:"id"`
		Attributes struct {
			ScriptName           string `json:"script_name"`
			JobName              string `json:"job_name"`
			Content              string `json:"content"`
			JobID                string `json:"job_id"`
			VariableSupport      bool   `json:"variable_support"`
			Status               string `json:"status"`
			PendingCount         int64  `json:"pending_count"`
			SuccessCount         int64  `json:"success_count"`
			ErroredCount         int64  `json:"errored_count"`
			CustomAttributeRegex string `json:"custom_attribute_regex"`
			CreatedBy            string `json:"created_by"`
			CreatedAt            string `json:"created_at"`
			UpdatedAt            string `json:"updated_at"`
		} `json:"attributes"`
		Relationships struct {
			CustomAttribute struct {
				Data *struct {
					ID string `json:"id"`
				} `json:"data"`
			} `json:"custom_attribute"`
			Device struct {
				Data []struct {
					ID         int     `json:"id"`
					Status     string  `json:"status"`
					StatusCode *string `json:"status_code"`
					Response   *string `json:"response"`
				} `json:"data"`
			} `json:"device"`
		} `json:"relationships"`
	} `json:"data"`
}

func fetchScriptJobDetails(ctx context.Context, client *simplemdm.Client, id string) (*scriptJobDetailsData, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s/api/v1/script_jobs/%s", client.HostName, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := client.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	var payload scriptJobDetailsResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	details := &scriptJobDetailsData{
		ID:                   strconv.Itoa(payload.Data.ID),
		ScriptName:           payload.Data.Attributes.ScriptName,
		JobName:              payload.Data.Attributes.JobName,
		JobIdentifier:        payload.Data.Attributes.JobID,
		Status:               payload.Data.Attributes.Status,
		PendingCount:         payload.Data.Attributes.PendingCount,
		SuccessCount:         payload.Data.Attributes.SuccessCount,
		ErroredCount:         payload.Data.Attributes.ErroredCount,
		Content:              payload.Data.Attributes.Content,
		VariableSupport:      payload.Data.Attributes.VariableSupport,
		CreatedBy:            payload.Data.Attributes.CreatedBy,
		CreatedAt:            payload.Data.Attributes.CreatedAt,
		UpdatedAt:            payload.Data.Attributes.UpdatedAt,
		CustomAttributeRegex: payload.Data.Attributes.CustomAttributeRegex,
	}

	if payload.Data.Relationships.CustomAttribute.Data != nil {
		details.CustomAttribute = payload.Data.Relationships.CustomAttribute.Data.ID
	}

	for _, device := range payload.Data.Relationships.Device.Data {
		deviceDetail := scriptJobDeviceDetail{
			ID:     strconv.Itoa(device.ID),
			Status: device.Status,
		}

		if device.StatusCode != nil && *device.StatusCode != "" {
			statusCode := *device.StatusCode
			deviceDetail.StatusCode = &statusCode
		}

		if device.Response != nil && *device.Response != "" {
			response := *device.Response
			deviceDetail.Response = &response
		}

		details.Devices = append(details.Devices, deviceDetail)
	}

	return details, nil
}

func scriptJobDevicesListValue(ctx context.Context, devices []scriptJobDeviceDetail) (types.List, diag.Diagnostics) {
	if len(devices) == 0 {
		return types.ListValue(types.ObjectType{AttrTypes: scriptJobDeviceAttrTypes}, []attr.Value{})
	}

	values := make([]attr.Value, 0, len(devices))
	var diags diag.Diagnostics

	for _, device := range devices {
		attrs := map[string]attr.Value{
			"id":     types.StringValue(device.ID),
			"status": types.StringValue(device.Status),
		}

		if device.StatusCode != nil {
			attrs["status_code"] = types.StringValue(*device.StatusCode)
		} else {
			attrs["status_code"] = types.StringNull()
		}

		if device.Response != nil {
			attrs["response"] = types.StringValue(*device.Response)
		} else {
			attrs["response"] = types.StringNull()
		}

		obj, d := types.ObjectValue(scriptJobDeviceAttrTypes, attrs)
		diags.Append(d...)
		values = append(values, obj)
	}

	list, d := types.ListValue(types.ObjectType{AttrTypes: scriptJobDeviceAttrTypes}, values)
	diags.Append(d...)

	return list, diags
}

type scriptJobResponse struct {
	Data scriptJobData `json:"data"`
}

type scriptJobData struct {
	ID            int                    `json:"id"`
	Type          string                 `json:"type"`
	Attributes    scriptJobAttributes    `json:"attributes"`
	Relationships scriptJobRelationships `json:"relationships"`
}

type scriptJobAttributes struct {
	JobName   string `json:"job_name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type scriptJobRelationships struct {
	Script          scriptJobRelationshipItem `json:"script"`
	AssignmentGroup scriptJobRelationshipItem `json:"assignment_group"`
}

type scriptJobRelationshipItem struct {
	Data *struct {
		ID   int    `json:"id"`
		Type string `json:"type"`
	} `json:"data"`
}

type scriptJobFlat struct {
	ID                  int
	JobName             string
	ScriptID            *string
	AssignmentGroupID   *string
	AssignmentGroupName string
	Status              string
	CreatedAt           string
	UpdatedAt           string
}

func listScriptJobs(ctx context.Context, client *simplemdm.Client, startingAfter int) ([]scriptJobResponse, error) {
	var allJobs []scriptJobResponse
	limit := 100

	for {
		url := fmt.Sprintf("https://%s/api/v1/script_jobs?limit=%d", client.HostName, limit)
		if startingAfter > 0 {
			url += fmt.Sprintf("&starting_after=%d", startingAfter)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		body, err := client.RequestResponse200(req)
		if err != nil {
			return nil, err
		}

		page, hasMore, err := simplemdm.DecodeList[scriptJobData](body)
		if err != nil {
			return nil, err
		}

		for _, data := range page {
			allJobs = append(allJobs, scriptJobResponse{Data: data})
		}

		if !hasMore {
			break
		}

		if len(page) > 0 {
			startingAfter = page[len(page)-1].ID
		} else {
			break
		}
	}

	return allJobs, nil
}

func flattenScriptJob(response *scriptJobResponse) scriptJobFlat {
	flat := scriptJobFlat{
		ID:        response.Data.ID,
		JobName:   response.Data.Attributes.JobName,
		Status:    response.Data.Attributes.Status,
		CreatedAt: response.Data.Attributes.CreatedAt,
		UpdatedAt: response.Data.Attributes.UpdatedAt,
	}

	if response.Data.Relationships.Script.Data != nil {
		scriptID := strconv.Itoa(response.Data.Relationships.Script.Data.ID)
		flat.ScriptID = &scriptID
	}

	if response.Data.Relationships.AssignmentGroup.Data != nil {
		groupID := strconv.Itoa(response.Data.Relationships.AssignmentGroup.Data.ID)
		flat.AssignmentGroupID = &groupID
		// Note: We don't have the name in the list response, so it will be empty
		flat.AssignmentGroupName = ""
	}

	return flat
}
