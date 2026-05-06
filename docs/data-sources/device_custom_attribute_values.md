---
page_title: "simplemdm_device_custom_attribute_values Data Source - simplemdm"
subcategory: ""
description: |-
  Lists custom attribute values resolved for a device.
---

# simplemdm_device_custom_attribute_values (Data Source)

Lists custom attribute values resolved for a device, including values inherited from groups or the account default. Each entry surfaces the resolved `value`, whether the attribute is `secret`, and the `source` of the value.

## Example Usage

```terraform
data "simplemdm_device_custom_attribute_values" "device42" {
  device_id = "42"
}
```

## Schema

### Required

- `device_id` (String) ID of the device.

### Read-Only (Block list)

Each `custom_attribute_values` block exposes:

- `id` (String) Custom attribute name.
- `value` (String)
- `secret` (Bool)
- `source` (String) `device`, `group`, or `account`.
