---
page_title: "simplemdm_dep_device Data Source - simplemdm"
subcategory: ""
description: |-
  Retrieves a single DEP device under a DEP server by its DEP device ID.
---

# simplemdm_dep_device (Data Source)

Retrieves a single DEP device under a DEP server by its DEP device ID.

## Example Usage

```terraform
data "simplemdm_dep_device" "specific" {
  dep_server_id = "1"
  id            = "42"
}
```

## Schema

### Required

- `dep_server_id` (String)
- `id` (String) DEP device ID.

### Read-Only

- `serial_number`, `model`, `color`, `description`, `os`, `device_family`, `profile_status`,
  `profile_assign_time`, `profile_push_time`, `device_assigned_date`, `device_assigned_by` (String).
