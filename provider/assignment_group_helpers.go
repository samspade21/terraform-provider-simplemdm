package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type assignmentGroupResponse struct {
	Data struct {
		ID            int                          `json:"id"`
		Type          string                       `json:"type"`
		Attributes    assignmentGroupAttributes    `json:"attributes"`
		Relationships assignmentGroupRelationships `json:"relationships"`
	} `json:"data"`
}

type assignmentGroupAttributes struct {
	Name             string `json:"name"`
	AutoDeploy       bool   `json:"auto_deploy"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	DeviceCount      int    `json:"device_count"`
	GroupCount       int    `json:"group_count"`
	Priority         *int   `json:"priority,omitempty"`
	AppTrackLocation *bool  `json:"app_track_location,omitempty"`
}

type assignmentGroupRelationships struct {
	Apps         assignmentGroupRelationshipItems `json:"apps"`
	Profiles     assignmentGroupRelationshipItems `json:"profiles"`
	Devices      assignmentGroupRelationshipItems `json:"devices"`
	DeviceGroups assignmentGroupRelationshipItems `json:"device_groups"`
}

type assignmentGroupRelationshipItems struct {
	Data []assignmentGroupRelationshipItem `json:"data"`
}

type assignmentGroupRelationshipItem struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

func fetchAssignmentGroup(ctx context.Context, client *simplemdm.Client, id string) (*assignmentGroupResponse, error) {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s", client.HostName, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := client.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	var assignmentGroup assignmentGroupResponse
	if err := json.Unmarshal(body, &assignmentGroup); err != nil {
		return nil, err
	}

	return &assignmentGroup, nil
}

func buildStringSetFromRelationshipItems(items []assignmentGroupRelationshipItem) types.Set {
	// Return empty set instead of null for Optional+Computed attributes
	// This prevents "was X but now null" errors when API doesn't return relationships
	values := make([]attr.Value, 0, len(items))
	for _, item := range items {
		values = append(values, types.StringValue(strconv.Itoa(item.ID)))
	}

	return types.SetValueMust(types.StringType, values)
}

type assignmentGroupUpsertRequest struct {
	Name             string
	AutoDeploy       *bool
	Priority         *int64
	AppTrackLocation *bool
}

func createAssignmentGroup(ctx context.Context, client *simplemdm.Client, payload assignmentGroupUpsertRequest) (*assignmentGroupResponse, error) {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups", client.HostName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = buildAssignmentGroupQuery(payload, true).Encode()

	body, err := client.RequestResponse201(req)
	if err != nil {
		return nil, err
	}

	var assignmentGroup assignmentGroupResponse
	if err := json.Unmarshal(body, &assignmentGroup); err != nil {
		return nil, fmt.Errorf("decode assignment group create response: %w (body=%s)", err, string(body))
	}

	if assignmentGroup.Data.ID != 0 {
		return &assignmentGroup, nil
	}

	// SimpleMDM's New Groups Experience accounts return `"id": null` in the
	// POST /assignment_groups response — the create succeeded but the body
	// can't tell us the new ID. Recover by listing /assignment_groups and
	// matching on name. The list is sorted newest-first, so we take the
	// first match (most-recently created group with that name).
	id, listErr := lookupAssignmentGroupIDByName(ctx, client, payload.Name)
	if listErr != nil {
		return nil, fmt.Errorf("create returned id=null and follow-up list lookup failed: %w (create body=%s)", listErr, string(body))
	}
	if id == 0 {
		return nil, fmt.Errorf("assignment group %q created but cannot be found in /assignment_groups list (create body=%s)", payload.Name, string(body))
	}
	assignmentGroup.Data.ID = id
	return &assignmentGroup, nil
}

// lookupAssignmentGroupIDByName fetches /assignment_groups (filtered with
// `search=<name>`) and returns the ID of the most-recently-created group
// with the given exact name. Used to recover from POST responses that omit
// the new ID (New Groups Experience accounts).
func lookupAssignmentGroupIDByName(ctx context.Context, client *simplemdm.Client, name string) (int, error) {
	endpoint := fmt.Sprintf("https://%s/api/v1/assignment_groups", client.HostName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}
	q := req.URL.Query()
	q.Set("search", name)
	q.Set("limit", "100")
	req.URL.RawQuery = q.Encode()

	body, err := client.RequestResponse200(req)
	if err != nil {
		return 0, err
	}
	var resp struct {
		Data []struct {
			ID         int `json:"id"`
			Attributes struct {
				Name string `json:"name"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, err
	}
	// Prefer an exact-name match (search is fuzzy).
	for _, g := range resp.Data {
		if g.Attributes.Name == name && g.ID != 0 {
			return g.ID, nil
		}
	}
	return 0, nil
}

func updateAssignmentGroup(ctx context.Context, client *simplemdm.Client, id string, payload assignmentGroupUpsertRequest) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s", client.HostName, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, nil)
	if err != nil {
		return err
	}

	req.URL.RawQuery = buildAssignmentGroupQuery(payload, false).Encode()

	_, err = client.RequestResponse204(req)
	return err
}

// buildAssignmentGroupQuery constructs the query parameters shared by the create and update operations.
// When includeName is true the "name" parameter is always sent, mirroring the API requirement for creation requests.
// For updates, the name is only provided when it has a non-empty value so partial updates remain possible.
func buildAssignmentGroupQuery(payload assignmentGroupUpsertRequest, includeName bool) url.Values {
	values := url.Values{}

	if includeName || payload.Name != "" {
		values.Set("name", payload.Name)
	}

	setOptionalBool(values, "auto_deploy", payload.AutoDeploy)

	if payload.Priority != nil {
		values.Set("priority", strconv.FormatInt(*payload.Priority, 10))
	}

	setOptionalBool(values, "app_track_location", payload.AppTrackLocation)

	return values
}

// setOptionalBool adds the given key to the query values when the pointer contains a value.
func setOptionalBool(values url.Values, key string, value *bool) {
	if value != nil {
		values.Set(key, strconv.FormatBool(*value))
	}
}

// setOptionalString adds the given key to the query values when the pointer contains a non-empty string.
func setOptionalString(values url.Values, key string, value *string) {
	if value != nil && *value != "" {
		values.Set(key, *value)
	}
}

func assignmentGroupAssignDevice(ctx context.Context, client *simplemdm.Client, groupID string, deviceID string, removeOthers bool) error {
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s/devices/%s", client.HostName, groupID, deviceID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	if removeOthers {
		q := req.URL.Query()
		q.Add("remove_others", "true")
		req.URL.RawQuery = q.Encode()
	}

	_, err = client.RequestResponse204(req)
	return err
}

func applyAssignmentGroupResponseToResourceModel(model *assignment_groupResourceModel, response *assignmentGroupResponse) {
	model.ID = types.StringValue(strconv.Itoa(response.Data.ID))
	model.Apps = buildStringSetFromRelationshipItems(response.Data.Relationships.Apps.Data)
	model.Groups = buildStringSetFromRelationshipItems(response.Data.Relationships.DeviceGroups.Data)
	model.Devices = buildStringSetFromRelationshipItems(response.Data.Relationships.Devices.Data)
	model.Profiles = buildStringSetFromRelationshipItems(response.Data.Relationships.Profiles.Data)

	model.Name = types.StringValue(response.Data.Attributes.Name)
	model.AutoDeploy = types.BoolValue(response.Data.Attributes.AutoDeploy)

	if response.Data.Attributes.Priority != nil {
		model.Priority = types.Int64Value(int64(*response.Data.Attributes.Priority))
	} else {
		model.Priority = types.Int64Null()
	}

	if response.Data.Attributes.AppTrackLocation != nil {
		model.AppTrackLocation = types.BoolValue(*response.Data.Attributes.AppTrackLocation)
	} else {
		model.AppTrackLocation = types.BoolNull()
	}

	model.CreatedAt = types.StringValue(response.Data.Attributes.CreatedAt)
	model.UpdatedAt = types.StringValue(response.Data.Attributes.UpdatedAt)

	model.DeviceCount = types.Int64Value(int64(response.Data.Attributes.DeviceCount))
	model.GroupCount = types.Int64Value(int64(response.Data.Attributes.GroupCount))
}

func applyAssignmentGroupResponseToDataSourceModel(model *assignmentGroupDataSourceModel, response *assignmentGroupResponse) {
	model.ID = types.StringValue(strconv.Itoa(response.Data.ID))
	model.Name = types.StringValue(response.Data.Attributes.Name)
	model.AutoDeploy = types.BoolValue(response.Data.Attributes.AutoDeploy)

	if response.Data.Attributes.Priority != nil {
		model.Priority = types.Int64Value(int64(*response.Data.Attributes.Priority))
	} else {
		model.Priority = types.Int64Null()
	}

	if response.Data.Attributes.AppTrackLocation != nil {
		model.AppTrackLocation = types.BoolValue(*response.Data.Attributes.AppTrackLocation)
	} else {
		model.AppTrackLocation = types.BoolNull()
	}

	// Always set CreatedAt and UpdatedAt, even if empty
	// This ensures Computed fields in data sources are considered "set"
	model.CreatedAt = types.StringValue(response.Data.Attributes.CreatedAt)
	model.UpdatedAt = types.StringValue(response.Data.Attributes.UpdatedAt)

	model.Apps = buildStringSetFromRelationshipItems(response.Data.Relationships.Apps.Data)
	model.Groups = buildStringSetFromRelationshipItems(response.Data.Relationships.DeviceGroups.Data)
	model.Devices = buildStringSetFromRelationshipItems(response.Data.Relationships.Devices.Data)
	model.Profiles = buildStringSetFromRelationshipItems(response.Data.Relationships.Profiles.Data)

	model.DeviceCount = types.Int64Value(int64(response.Data.Attributes.DeviceCount))
	model.GroupCount = types.Int64Value(int64(response.Data.Attributes.GroupCount))
}

// setElementsToStringSlice converts a types.Set to a []string slice
func setElementsToStringSlice(set types.Set) []string {
	if set.IsNull() || set.IsUnknown() {
		return []string{}
	}

	elements := set.Elements()
	result := make([]string, 0, len(elements))
	for _, element := range elements {
		stringElement, ok := element.(types.String)
		if !ok || stringElement.IsNull() || stringElement.IsUnknown() {
			continue
		}

		result = append(result, stringElement.ValueString())
	}
	return result
}

// assignObjectsToGroup assigns multiple objects to an assignment group.
// Used during Create operations to assign profiles, groups, or devices.
// For apps see assignAppsToGroup which handles the per-app parameter
// overrides (install_type / deployment_type).
//
// SimpleMDM has a brief eventual-consistency window between when a freshly
// created assignment group becomes addressable by ID and when its associated
// /assignment_groups/{id}/<rel>/{otherID} endpoints start responding. The
// first call after Create occasionally 404s for a few seconds before
// settling. Retrying with backoff is enough to ride out the window without
// slowing down the happy path.
func assignObjectsToGroup(
	ctx context.Context,
	client *simplemdm.Client,
	groupID string,
	objects types.Set,
	objectType string,
	removeOthers bool,
) error {
	if objects.IsNull() || objects.IsUnknown() {
		return nil
	}

	for _, objectID := range objects.Elements() {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("operation cancelled: %w", err)
		}

		idString := objectID.(types.String).ValueString()

		assign := func() error {
			if objectType == "devices" {
				return assignmentGroupAssignDevice(ctx, client, groupID, idString, removeOthers)
			}
			return client.AssignmentGroupAssignObject(groupID, idString, objectType)
		}

		if err := retryOn404(ctx, assign); err != nil {
			return err
		}
	}
	return nil
}

// assignAppsToGroup is the apps-specific version of assignObjectsToGroup
// that goes through the SimpleMDM "Assign App" endpoint
// (POST /api/v1/assignment_groups/{id}/apps/{app_id}) which accepts optional
// `deployment_type` and `install_type` query params. The per-app overrides
// come from the resource's `apps_deployment_types` / `apps_install_types`
// maps; apps without an entry are sent without overrides and SimpleMDM
// applies the defaults documented at
// https://api.simplemdm.com/v1#assign-app.
func assignAppsToGroup(
	ctx context.Context,
	client *simplemdm.Client,
	groupID string,
	apps types.Set,
	installTypes map[string]string,
	deploymentTypes map[string]string,
) error {
	if apps.IsNull() || apps.IsUnknown() {
		return nil
	}

	for _, appElem := range apps.Elements() {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("operation cancelled: %w", err)
		}
		appID := appElem.(types.String).ValueString()
		install := installTypes[appID]
		deployment := deploymentTypes[appID]

		assign := func() error {
			return client.AssignmentGroupAssignApp(groupID, appID, deployment, install)
		}
		if err := retryOn404(ctx, assign); err != nil {
			return err
		}
	}
	return nil
}

// stringMapFromTypesMap converts a types.Map of strings into a Go map.
// Returns an empty map when the input is null / unknown so callers can use
// `m[key]` safely.
func stringMapFromTypesMap(m types.Map) map[string]string {
	if m.IsNull() || m.IsUnknown() {
		return map[string]string{}
	}
	out := make(map[string]string, len(m.Elements()))
	for k, v := range m.Elements() {
		if str, ok := v.(types.String); ok {
			out[k] = str.ValueString()
		}
	}
	return out
}

// retryOn404 calls fn and, if it returns a 404 error, retries with
// exponential backoff. Returns the last error if all attempts fail. Other
// error types are returned immediately.
//
// Used to ride out the eventual-consistency window between Create and the
// associated relationship endpoints becoming addressable. Total wait
// across all retries is ~30s (1+2+4+8+15s) which we found empirically
// to cover SimpleMDM's worst-case propagation delay on freshly-created
// assignment groups.
func retryOn404(ctx context.Context, fn func() error) error {
	delays := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		15 * time.Second,
	}
	var lastErr error
	for attempt := 0; attempt <= len(delays); attempt++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("operation cancelled: %w", err)
		}
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if !isNotFoundError(lastErr) || attempt == len(delays) {
			return lastErr
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		case <-time.After(delays[attempt]):
		}
	}
	return lastErr
}

// waitForAssignmentGroupAddressable blocks until /assignment_groups/{id}
// returns 200, or times out. Called after Create to ride out the
// eventual-consistency window before any per-relationship assign calls.
// Empirically, GETting the group as soon as it becomes addressable is a
// reliable signal that the relationship sub-resources are also addressable.
func waitForAssignmentGroupAddressable(ctx context.Context, client *simplemdm.Client, groupID string) error {
	return retryOn404(ctx, func() error {
		_, err := fetchAssignmentGroup(ctx, client, groupID)
		return err
	})
}

// updateAssignmentGroupApps reconciles the apps assigned to an assignment
// group between state and plan. New apps go through AssignmentGroupAssignApp
// so per-app install_type / deployment_type overrides apply; removed apps
// go through AssignmentGroupUnAssignApp.
//
// Apps that stay in the set but whose install_type or deployment_type
// changed are re-assigned (the API treats this as idempotent — it just
// updates the existing relationship).
func updateAssignmentGroupApps(
	ctx context.Context,
	client *simplemdm.Client,
	groupID string,
	stateApps, planApps types.Set,
	stateInstall, planInstall map[string]string,
	stateDeploy, planDeploy map[string]string,
) error {
	if planApps.IsNull() || planApps.IsUnknown() {
		return nil
	}
	if stateApps.IsNull() || stateApps.IsUnknown() {
		stateApps = types.SetNull(types.StringType)
	}

	stateSlice := setElementsToStringSlice(stateApps)
	planSlice := setElementsToStringSlice(planApps)
	toAdd, toRemove := diffFunction(stateSlice, planSlice)

	for _, appID := range toAdd {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("operation cancelled: %w", err)
		}
		if err := client.AssignmentGroupAssignApp(groupID, appID, planDeploy[appID], planInstall[appID]); err != nil {
			return err
		}
	}

	// Re-assign apps whose overrides changed in the new plan.
	for _, appID := range planSlice {
		if !slices.Contains(stateSlice, appID) {
			continue // already added above
		}
		if planInstall[appID] != stateInstall[appID] || planDeploy[appID] != stateDeploy[appID] {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("operation cancelled: %w", err)
			}
			if err := client.AssignmentGroupAssignApp(groupID, appID, planDeploy[appID], planInstall[appID]); err != nil {
				return err
			}
		}
	}

	for _, appID := range toRemove {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("operation cancelled: %w", err)
		}
		if err := client.AssignmentGroupUnAssignApp(groupID, appID); err != nil {
			return err
		}
	}
	return nil
}

// updateAssignmentGroupObjects updates assignments by computing diff and applying changes
// Used during Update operations to sync state with plan
func updateAssignmentGroupObjects(
	ctx context.Context,
	client *simplemdm.Client,
	groupID string,
	stateObjects types.Set,
	planObjects types.Set,
	objectType string,
	removeOthers bool,
) error {
	// When the plan set is null or unknown, Terraform cannot determine a desired
	// target state for the relationship. In these cases we must leave the
	// existing assignments untouched and bail out early regardless of what the
	// current state contains.
	if planObjects.IsNull() || planObjects.IsUnknown() {
		return nil
	}

	if stateObjects.IsNull() || stateObjects.IsUnknown() {
		stateObjects = types.SetNull(types.StringType)
	}

	// Convert sets to string slices
	stateSlice := setElementsToStringSlice(stateObjects)
	planSlice := setElementsToStringSlice(planObjects)

	// Compute diff
	toAdd, toRemove := diffFunction(stateSlice, planSlice)

	// Add new objects
	for _, objectID := range toAdd {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("operation cancelled: %w", err)
		}

		var err error
		if objectType == "devices" {
			err = assignmentGroupAssignDevice(ctx, client, groupID, objectID, removeOthers)
		} else {
			err = client.AssignmentGroupAssignObject(groupID, objectID, objectType)
		}
		if err != nil {
			return err
		}
	}

	// Remove old objects
	for _, objectID := range toRemove {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("operation cancelled: %w", err)
		}

		err := client.AssignmentGroupUnAssignObject(groupID, objectID, objectType)
		if err != nil {
			return err
		}
	}

	return nil
}

// diffFunction computes the difference between state and plan lists
// Returns items to add and items to remove
func diffFunction(state []string, plan []string) (add []string, remove []string) {
	// Create map of state items for O(1) lookups
	stateMap := make(map[string]bool, len(state))
	for _, s := range state {
		stateMap[s] = true
	}

	// Create map of plan items and identify additions
	planMap := make(map[string]bool, len(plan))
	for _, p := range plan {
		planMap[p] = true
		if !stateMap[p] {
			add = append(add, p)
		}
	}

	// Identify removals
	for _, s := range state {
		if !planMap[s] {
			remove = append(remove, s)
		}
	}

	return add, remove
}

// preservePlannedRelationships handles eventual consistency by preserving planned values
// when the API doesn't immediately return assigned relationships
func preservePlannedRelationships(
	model *assignment_groupResourceModel,
	plannedApps, plannedProfiles, plannedGroups, plannedDevices types.Set,
	apiReturnedApps, apiReturnedProfiles, apiReturnedGroups, apiReturnedDevices bool,
) {
	// Restore planned relationship values if they were set but API returned empty
	// This prevents "planned X but got Y" errors due to API eventual consistency
	if !plannedApps.IsNull() && !plannedApps.IsUnknown() && !apiReturnedApps {
		model.Apps = plannedApps
	}
	if !plannedProfiles.IsNull() && !plannedProfiles.IsUnknown() && !apiReturnedProfiles {
		model.Profiles = plannedProfiles
	}
	if !plannedGroups.IsNull() && !plannedGroups.IsUnknown() && !apiReturnedGroups {
		model.Groups = plannedGroups
	}
	if !plannedDevices.IsNull() && !plannedDevices.IsUnknown() && !apiReturnedDevices {
		model.Devices = plannedDevices
	}
}
