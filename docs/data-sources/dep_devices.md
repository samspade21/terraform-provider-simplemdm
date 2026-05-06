---
page_title: "simplemdm_dep_devices Data Source - simplemdm"
subcategory: ""
description: |-
  Lists DEP devices reported by the given Apple DEP server.
---

# simplemdm_dep_devices (Data Source)

Lists DEP devices reported by the given Apple DEP server.

## Example Usage

```terraform
data "simplemdm_dep_devices" "all" {
  dep_server_id = "1"
}
```

## Schema

### Required

- `dep_server_id` (String) DEP server ID whose devices to list.

### Read-Only (Block list)

Each `dep_devices` block exposes:

- `id`, `serial_number`, `model`, `color`, `description`, `os`, `device_family`, `profile_status`,
  `profile_assign_time`, `profile_push_time`, `device_assigned_date`, `device_assigned_by` (String).
