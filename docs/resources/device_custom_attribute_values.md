---
page_title: "simplemdm_device_custom_attribute_values Resource - simplemdm"
subcategory: ""
description: |-
  Sets multiple custom attribute values on a single device in one PUT call.
---

# simplemdm_device_custom_attribute_values (Resource)

Sets multiple custom attribute values on a single device in one PUT call. This is a fire-and-apply resource: changes to `assignments` apply on Update; the resource has no per-attribute drift detection. Use `triggers` to force a re-apply.

## Example Usage

```terraform
resource "simplemdm_device_custom_attribute_values" "alice_phone" {
  device_id = "1234"

  assignments {
    name  = "department"
    value = "platform-team"
  }

  assignments {
    name  = "office"
    value = "berlin"
  }
}
```

## Schema

### Required

- `device_id` (String)
- `assignments` (Block list, min: 1)

#### `assignments` block

- `name` (String, Required)
- `value` (String, Required)

### Optional

- `triggers` (Map of String) Changing any value forces a replace and re-applies the bulk update.

### Read-Only

- `id` (String) Synthetic identifier (`device_cav:<device_id>`).
- `last_applied` (String)
