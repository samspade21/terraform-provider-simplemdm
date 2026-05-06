---
page_title: "simplemdm_device_custom_attribute_value Resource - simplemdm"
subcategory: ""
description: |-
  Sets a custom attribute value on a specific device.
---

# simplemdm_device_custom_attribute_value (Resource)

Sets a custom attribute value on a specific device. Deleting the resource clears the value (sets it to empty string).

## Example Usage

```terraform
resource "simplemdm_attribute" "department" {
  name          = "department"
  default_value = "engineering"
}

resource "simplemdm_device_custom_attribute_value" "alice" {
  device_id      = "1234"
  attribute_name = simplemdm_attribute.department.name
  value          = "platform-team"
}
```

## Schema

### Required

- `device_id` (String)
- `attribute_name` (String) Custom attribute name (must already exist).
- `value` (String)

### Read-Only

- `id` (String) Composite ID `<device_id>/<attribute_name>`.
