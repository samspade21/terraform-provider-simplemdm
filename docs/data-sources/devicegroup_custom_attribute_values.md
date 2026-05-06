---
page_title: "simplemdm_devicegroup_custom_attribute_values Data Source - simplemdm"
subcategory: ""
description: |-
  Lists custom attribute values assigned to a (legacy) device group.
---

# simplemdm_devicegroup_custom_attribute_values (Data Source)

Lists custom attribute values assigned to a (legacy) device group. Device groups are deprecated in favor of assignment groups; new deployments should prefer `simplemdm_assignmentgroup_custom_attribute_values`.

## Example Usage

```terraform
data "simplemdm_devicegroup_custom_attribute_values" "group" {
  device_group_id = "1"
}
```

## Schema

### Required

- `device_group_id` (String) ID of the device group.

### Read-Only (Block list)

Each `custom_attribute_values` block exposes `id`, `value`, `secret`, `source`.
