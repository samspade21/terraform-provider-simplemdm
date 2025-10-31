---
page_title: "simplemdm_customprofiles Data Source - terraform-provider-simplemdm"
subcategory: ""
description: |-
  Fetches the collection of custom configuration profiles from your SimpleMDM account.
---

# simplemdm_customprofiles (Data Source)

Fetches the collection of custom configuration profiles from your SimpleMDM account.

## Example Usage

```terraform
data "simplemdm_customprofiles" "all" {
}

output "custom_profile_count" {
  value = length(data.simplemdm_customprofiles.all.custom_profiles)
}

output "custom_profile_names" {
  value = [for profile in data.simplemdm_customprofiles.all.custom_profiles : profile.name]
}
```

## Schema

### Read-Only

- `custom_profiles` (Block List) Collection of custom profile records returned by the API. (see [below for nested schema](#nestedblock--custom_profiles))

<a id="nestedblock--custom_profiles"></a>
### Nested Schema for `custom_profiles`

Read-Only:

- `attributesupport` (Boolean) Indicates whether variable substitution is enabled for the profile.
- `devicecount` (Number) Number of devices currently assigned to this profile.
- `escapeattributes` (Boolean) Indicates whether custom attribute values are escaped when substituted into the profile.
- `groupcount` (Number) Number of device groups currently assigned to this profile.
- `id` (String) Custom profile identifier.
- `name` (String) The name of the custom profile.
- `profileidentifier` (String) Profile identifier assigned by SimpleMDM.
- `reinstallafterosupdate` (Boolean) Whether the profile reinstalls automatically after macOS updates.
- `userscope` (Boolean) Whether the profile deploys as a user profile for macOS devices.