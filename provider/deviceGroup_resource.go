package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/DavidKrau/simplemdm-go-client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &deviceGroupResource{}
	_ resource.ResourceWithConfigure   = &deviceGroupResource{}
	_ resource.ResourceWithImportState = &deviceGroupResource{}
)

// deviceGroupResourceModel maps the resource schema data.
type deviceGroupResourceModel struct {
	Name           types.String `tfsdk:"name"`
	ID             types.String `tfsdk:"id"`
	Attributes     types.Map    `tfsdk:"attributes"`
	Profiles       types.Set    `tfsdk:"profiles"`
	CustomProfiles types.Set    `tfsdk:"customprofiles"`
	CloneFrom      types.String `tfsdk:"clone_from"`
}

// deviceGroupResource is a helper function to simplify the provider implementation.
func DeviceGroupResource() resource.Resource {
	return &deviceGroupResource{}
}

// deviceGroupResource is the resource implementation.
type deviceGroupResource struct {
	client *simplemdm.Client
}

// Configure adds the provider configured client to the resource.
func (r *deviceGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*simplemdm.Client)
}

// Metadata returns the resource type name.
func (r *deviceGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_devicegroup"
}

// Schema defines the schema for the resource.
func (r *deviceGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Device Group resource can be used to manage Device Group. Can be used together with Custom Profile(s), Attribute(s), Assignment Group(s) or Device Group(s) and set addition details regarding Device Group.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Optional:    false,
				Description: "Required. The name of the device group.",
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "ID of a Device Group in SimpleMDM",
			},
			"profiles": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Optional. List of Configuration Profiles assigned to this Device Group",
			},
			"customprofiles": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Optional. List of Custom Configuration Profiles assigned to this Device Group",
			},
			"clone_from": schema.StringAttribute{
				Optional:    true,
				Description: "Optional. Clone configuration from an existing legacy device group. Changing this value forces a new device group to be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"attributes": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Optional. Map of Custom Configuration Profiles and values set for this Device Group",
			},
		},
	}
}

func (r *deviceGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to state
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create a new resource
func (r *deviceGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	//Retrieve values from plan
	var plan deviceGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var (
		deviceGroup *simplemdm.SimplemdmDefaultStruct
		err         error
	)

	if !plan.CloneFrom.IsNull() && !plan.CloneFrom.IsUnknown() && plan.CloneFrom.ValueString() != "" {
		deviceGroup, err = cloneDeviceGroup(ctx, r.client, plan.CloneFrom.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error cloning device group",
				"Could not clone SimpleMDM device group: "+err.Error(),
			)
			return
		}
	} else {
		deviceGroup, err = r.client.DeviceGroupCreate(plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating device group",
				"Could not create SimpleMDM device group: "+err.Error(),
			)
			return
		}
	}

	plan.ID = types.StringValue(strconv.Itoa(deviceGroup.Data.ID))

	if plan.Name.ValueString() != "" && plan.Name.ValueString() != deviceGroup.Data.Attributes.Name {
		if err := updateDeviceGroupName(ctx, r.client, plan.ID.ValueString(), plan.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Error updating device group name",
				"Could not update SimpleMDM device group name: "+err.Error(),
			)
			return
		}
	}

	r.reconcileAttributes(ctx, plan.ID.ValueString(), types.MapNull(types.StringType), plan.Attributes, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.reconcileProfiles(ctx, plan.ID.ValueString(), types.SetNull(types.StringType), plan.Profiles, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.reconcileCustomProfiles(ctx, plan.ID.ValueString(), types.SetNull(types.StringType), plan.CustomProfiles, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *deviceGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state deviceGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get device group value from SimpleMDM
	devicegroup, err := r.client.DeviceGroupGet(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SimpleMDM device group",
			"Could not read SimpleMDM device group "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	//load attributes for given group
	attributes, err := r.client.AttributeGetAttributesForDeviceGroup(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SimpleMDM device group attributes",
			"Could not read SimpleMDM device group attributes"+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	//adding attributes to the map
	attributePresent := false
	attributesElements := map[string]attr.Value{}
	for _, attribute := range attributes.Data {
		if attribute.Attributes.Source == "group" {
			attributesElements[attribute.ID] = types.StringValue(attribute.Attributes.Value)
			attributePresent = true
		}
	}
	if attributePresent {
		attributesSetValue, _ := types.MapValue(types.StringType, attributesElements)
		state.Attributes = attributesSetValue
	} else {
		attributesSetValue := types.MapNull(types.StringType)
		state.Attributes = attributesSetValue
	}

	// Overwrite items with refreshed state
	state.Name = types.StringValue(devicegroup.Data.Attributes.Name)

	// Load all profiles in SimpleMDM
	profiles, err := r.client.ProfileGetAll()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SimpleMDM profiles",
			"Could not read SimpleMDM profiles: "+err.Error(),
		)
		return
	}
	// //read all profiles and put them to slice
	profilesPresent := false
	profilesElements := []attr.Value{}
	customProfilesPresent := false
	customProfilesElements := []attr.Value{}

	for _, profile := range profiles.Data { //<<edit here
		for _, group := range profile.Relationships.DeviceGroups.Data {
			if strconv.Itoa(group.ID) == state.ID.ValueString() {
				if profile.Type == "custom_configuration_profile" {
					customProfilesElements = append(customProfilesElements, types.StringValue(strconv.Itoa(profile.ID)))
					customProfilesPresent = true
				} else {
					profilesElements = append(profilesElements, types.StringValue(strconv.Itoa(profile.ID)))
					profilesPresent = true
				}
			}
		}

	}

	//if there are profile or custom profiles return them to state
	if profilesPresent {
		profilesSetValue, _ := types.SetValue(types.StringType, profilesElements)
		state.Profiles = profilesSetValue
	} else {
		profilesSetValue := types.SetNull(types.StringType)
		state.Profiles = profilesSetValue
	}

	if customProfilesPresent {
		customProfilesSetValue, _ := types.SetValue(types.StringType, customProfilesElements)
		state.CustomProfiles = customProfilesSetValue
	} else {
		customProfilesSetValue := types.SetNull(types.StringType)
		state.CustomProfiles = customProfilesSetValue
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *deviceGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//Retrieve values from plan
	var plan, state deviceGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Name.ValueString() != state.Name.ValueString() {
		if err := updateDeviceGroupName(ctx, r.client, plan.ID.ValueString(), plan.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Error updating device group name",
				"Could not update SimpleMDM device group name "+plan.ID.ValueString()+": "+err.Error(),
			)
			return
		}
	}

	r.reconcileAttributes(ctx, plan.ID.ValueString(), state.Attributes, plan.Attributes, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.reconcileProfiles(ctx, plan.ID.ValueString(), state.Profiles, plan.Profiles, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.reconcileCustomProfiles(ctx, plan.ID.ValueString(), state.CustomProfiles, plan.CustomProfiles, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *deviceGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deviceGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeviceGroupDelete(state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting SimpleMDM device group",
			"Could not delete SimpleMDM device group "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *deviceGroupResource) reconcileAttributes(ctx context.Context, groupID string, oldAttributes, newAttributes types.Map, diags *diag.Diagnostics) {
	_ = ctx

	if diags.HasError() {
		return
	}

	planElements := map[string]attr.Value{}
	stateElements := map[string]attr.Value{}

	if !newAttributes.IsNull() && !newAttributes.IsUnknown() {
		planElements = newAttributes.Elements()
	}

	if !oldAttributes.IsNull() && !oldAttributes.IsUnknown() {
		stateElements = oldAttributes.Elements()
	}

	for planAttribute, planValue := range planElements {
		trimmed := strings.Replace(planValue.String(), "\"", "", 2)

		if stateValue, ok := stateElements[planAttribute]; ok {
			if planValue.Equal(stateValue) {
				continue
			}
		}

		if err := r.client.AttributeSetAttributeForDeviceGroup(groupID, planAttribute, trimmed); err != nil {
			diags.AddError(
				"Error updating SimpleMDM device group attribute",
				fmt.Sprintf("Could not update attribute %q on device group %s: %s", planAttribute, groupID, err.Error()),
			)
			return
		}
	}

	for stateAttribute := range stateElements {
		if _, ok := planElements[stateAttribute]; ok {
			continue
		}

		if err := r.client.AttributeSetAttributeForDeviceGroup(groupID, stateAttribute, ""); err != nil {
			diags.AddError(
				"Error clearing SimpleMDM device group attribute",
				fmt.Sprintf("Could not clear attribute %q on device group %s: %s", stateAttribute, groupID, err.Error()),
			)
			return
		}
	}
}

func (r *deviceGroupResource) reconcileProfiles(ctx context.Context, groupID string, oldProfiles, newProfiles types.Set, diags *diag.Diagnostics) {
	_ = ctx

	if diags.HasError() {
		return
	}

	stateProfiles := extractStringSet(oldProfiles)
	planProfiles := extractStringSet(newProfiles)

	profilesToAdd, profilesToRemove := diffFunction(stateProfiles, planProfiles)

	for _, profileID := range profilesToAdd {
		if err := r.client.ProfileAssignToGroup(profileID, groupID); err != nil {
			diags.AddError(
				"Error assigning profile to device group",
				fmt.Sprintf("Could not assign profile %s to device group %s: %s", profileID, groupID, err.Error()),
			)
			return
		}
	}

	for _, profileID := range profilesToRemove {
		if err := r.client.ProfileUnAssignToGroup(profileID, groupID); err != nil {
			diags.AddError(
				"Error unassigning profile from device group",
				fmt.Sprintf("Could not unassign profile %s from device group %s: %s", profileID, groupID, err.Error()),
			)
			return
		}
	}
}

func (r *deviceGroupResource) reconcileCustomProfiles(ctx context.Context, groupID string, oldProfiles, newProfiles types.Set, diags *diag.Diagnostics) {
	_ = ctx

	if diags.HasError() {
		return
	}

	stateProfiles := extractStringSet(oldProfiles)
	planProfiles := extractStringSet(newProfiles)

	profilesToAdd, profilesToRemove := diffFunction(stateProfiles, planProfiles)

	for _, profileID := range profilesToAdd {
		if err := r.client.CustomProfileAssignToDeviceGroup(profileID, groupID); err != nil {
			diags.AddError(
				"Error assigning custom profile to device group",
				fmt.Sprintf("Could not assign custom profile %s to device group %s: %s", profileID, groupID, err.Error()),
			)
			return
		}
	}

	for _, profileID := range profilesToRemove {
		if err := r.client.CustomProfileUnassignFromDeviceGroup(profileID, groupID); err != nil {
			diags.AddError(
				"Error unassigning custom profile from device group",
				fmt.Sprintf("Could not unassign custom profile %s from device group %s: %s", profileID, groupID, err.Error()),
			)
			return
		}
	}
}

func extractStringSet(set types.Set) []string {
	if set.IsNull() || set.IsUnknown() {
		return []string{}
	}

	values := []string{}
	for _, value := range set.Elements() {
		values = append(values, strings.Replace(value.String(), "\"", "", 2))
	}

	return values
}

func updateDeviceGroupName(ctx context.Context, client *simplemdm.Client, groupID, name string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("https://%s/api/v1/device_groups/%s", client.HostName, groupID), nil)
	if err != nil {
		return err
	}

	query := req.URL.Query()
	query.Add("name", name)
	req.URL.RawQuery = query.Encode()

	_, err = client.RequestResponse200(req)
	return err
}

func cloneDeviceGroup(ctx context.Context, client *simplemdm.Client, sourceID string) (*simplemdm.SimplemdmDefaultStruct, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://%s/api/v1/device_groups/%s/clone", client.HostName, sourceID), nil)
	if err != nil {
		return nil, err
	}

	body, err := client.RequestResponse200(req)
	if err != nil {
		return nil, err
	}

	clonedGroup := simplemdm.SimplemdmDefaultStruct{}
	if err := json.Unmarshal(body, &clonedGroup); err != nil {
		return nil, err
	}

	return &clonedGroup, nil
}
