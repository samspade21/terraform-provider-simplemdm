---
page_title: "simplemdm_app_installs Data Source - simplemdm"
subcategory: ""
description: |-
  Lists every installation record (per device) of a managed application.
---

# simplemdm_app_installs (Data Source)

Lists every installation record (per device) of a managed application.

## Example Usage

```terraform
data "simplemdm_app_installs" "example" {
  app_id = "1234"
}

output "managed_devices" {
  value = [
    for install in data.simplemdm_app_installs.example.installs :
    install.device_id if install.managed
  ]
}
```

## Schema

### Required

- `app_id` (String) ID of the application whose install records to list.

### Read-Only (Block list)

Each `installs` block exposes:

- `id` (String) Installed app record ID.
- `name`, `identifier`, `version`, `short_version` (String)
- `bundle_size`, `dynamic_size` (Number, bytes)
- `managed` (Bool) Whether the install is managed by SimpleMDM.
- `discovered_at`, `last_seen_at` (String) RFC3339-style timestamps.
- `device_id` (String) ID of the device the application is installed on.
