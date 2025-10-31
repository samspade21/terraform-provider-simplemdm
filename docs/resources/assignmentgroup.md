---
page_title: "simplemdm_assignmentgroup Resource - terraform-provider-simplemdm"
subcategory: ""
description: |-
  Manages assignment groups in SimpleMDM, which are used to organize and deploy apps, profiles, and configurations to devices.
---

# simplemdm_assignmentgroup (Resource)

Manages assignment groups in SimpleMDM. Assignment groups allow you to organize devices and automatically deploy apps, custom profiles, and configuration profiles to those devices. This is the modern replacement for the deprecated Device Groups functionality.

## Example Usage

### Basic Example

```terraform
resource "simplemdm_assignmentgroup" "engineering" {
  name        = "Engineering Team"
  auto_deploy = true
}
```

### Advanced Example - Complete Configuration

```terraform
resource "simplemdm_assignmentgroup" "sales_team" {
  name        = "Sales Team"
  auto_deploy = true
  priority    = 10

  # App tracking
  app_track_location = true

  # Assign apps
  apps = [
    simplemdm_app.slack.id,
    simplemdm_app.salesforce.id,
  ]

  # Assign profiles
  profiles = [
    simplemdm_customprofile.vpn_config.id,
    simplemdm_customprofile.wifi_config.id,
  ]

  # Assign device groups
  groups = [
    simplemdm_devicegroup.mobile_devices.id,
  ]

  # Assign specific devices
  devices               = ["123456", "234567"]
  devices_remove_others = false

  # Post-operation commands
  profiles_sync = true
  apps_push     = true
  apps_update   = false
}
```

### Advanced Example - With Deprecated Fields

```terraform
# Example showing deprecated fields that may still be used for legacy compatibility
resource "simplemdm_assignmentgroup" "legacy_munki" {
  name = "Legacy Munki Group"

  # ⚠️ DEPRECATED: group_type is deprecated by SimpleMDM API
  # May be ignored for accounts using the New Groups Experience
  group_type = "munki"

  # ⚠️ DEPRECATED: install_type is deprecated by SimpleMDM API
  # SimpleMDM recommends setting install_type per-app instead
  install_type = "managed"

  auto_deploy = true
}
```

### Advanced Example - Multi-tier Deployment

```terraform
# High priority group for executives
resource "simplemdm_assignmentgroup" "executives" {
  name        = "Executive Devices"
  priority    = 1
  auto_deploy = true

  apps = [
    simplemdm_app.enterprise_suite.id,
  ]
}

# Standard group for general employees
resource "simplemdm_assignmentgroup" "employees" {
  name        = "All Employees"
  priority    = 10
  auto_deploy = true

  apps = [
    simplemdm_app.basic_tools.id,
  ]
}
```

## Schema

### Required

- `name` (String) The name of the assignment group.

### Optional

- `app_track_location` (Boolean) Controls whether the SimpleMDM app tracks device location when installed. Default: `false`.
- `apps` (Set of String) List of app IDs assigned to this assignment group.
- `apps_push` (Boolean) Send push apps command after assignment group creation or changes. Default: `false`.
- `apps_update` (Boolean) Send update apps command after assignment group creation or changes. Default: `false`.
- `auto_deploy` (Boolean) Whether apps should be automatically pushed to devices when they join this assignment group. Default: `true`.
- `devices` (Set of String) List of device IDs assigned to this assignment group.
- `devices_remove_others` (Boolean) When true, devices assigned through Terraform will be removed from other assignment groups before being added to this one. Default: `false`.
- `group_type` (String) Type of assignment group. Valid values: `standard` (for MDM app/media deployments) or `munki` (for Munki app deployments). Default: `standard`. **⚠️ DEPRECATED**: This field is deprecated by the SimpleMDM API and may be ignored for accounts using the New Groups Experience.
- `groups` (Set of String) List of device group IDs assigned to this assignment group.
- `install_type` (String) The install type for munki assignment groups. Valid values: `managed`, `self_serve`, `managed_updates`, `default_installs`. This setting has no effect for non-munki (standard) assignment groups. Default: `managed` for munki groups. **⚠️ DEPRECATED**: The SimpleMDM API recommends setting install_type per-app using the Assign App endpoint instead of at the group level.
- `priority` (Number) Sets the priority order in which assignment groups are evaluated when devices are part of multiple groups. Lower numbers have higher priority.
- `profiles` (Set of String) List of configuration profile IDs (both custom and predefined profiles) assigned to this assignment group.
- `profiles_sync` (Boolean) Send sync profiles command after assignment group creation or changes. Default: `false`.

### Read-Only

- `created_at` (String) Timestamp when the assignment group was created.
- `device_count` (Number) Number of devices currently assigned to the assignment group.
- `group_count` (Number) Number of device groups currently assigned to the assignment group.
- `id` (String) The unique identifier of the assignment group in SimpleMDM.
- `updated_at` (String) Timestamp when the assignment group was last updated.

## Import

Import is supported using the following syntax:

```shell
# Assignment group can be imported by specifying the assignment group ID
terraform import simplemdm_assignmentgroup.example 123456
```

## Notes

- **Priority Management**: When devices belong to multiple assignment groups, the group with the lowest `priority` value takes precedence.
- **Device Removal**: Use `devices_remove_others = true` carefully, as it will remove devices from all other assignment groups.
- **Post-Operation Commands**: The `profiles_sync`, `apps_push`, and `apps_update` flags trigger immediate actions after group changes. Use these when you need immediate deployment.
- **Deprecated Fields**: The `group_type` and `install_type` fields are maintained for backward compatibility but may be removed in future versions.
- **New Groups Experience**: Accounts using SimpleMDM's New Groups Experience should avoid using deprecated fields.