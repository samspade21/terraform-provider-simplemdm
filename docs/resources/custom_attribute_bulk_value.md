---
page_title: "simplemdm_custom_attribute_bulk_value Resource - simplemdm"
subcategory: ""
description: |-
  Pushes a custom attribute value across multiple devices in one PUT call.
---

# simplemdm_custom_attribute_bulk_value (Resource)

Pushes a custom attribute value across multiple devices in one PUT call. This is a fire-and-apply resource: changes to `assignments` apply on Update; the resource has no per-device drift detection. Use `triggers` to force a re-apply.

## Example Usage

```terraform
resource "simplemdm_custom_attribute_bulk_value" "office_assignments" {
  attribute_name = simplemdm_attribute.office.name

  assignments {
    device_id = "1234"
    value     = "berlin"
  }

  assignments {
    device_id = "5678"
    value     = "london"
  }

  triggers = {
    nonce = "2025-05-06"
  }
}
```

## Schema

### Required

- `attribute_name` (String)
- `assignments` (Block list, min: 1)

#### `assignments` block

- `device_id` (String, Required)
- `value` (String, Required)

### Optional

- `triggers` (Map of String) Changing any value forces a replace and re-applies the bulk update.

### Read-Only

- `id` (String) Synthetic identifier (`bulk:<attribute_name>`).
- `last_applied` (String)
