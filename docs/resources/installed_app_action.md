---
page_title: "simplemdm_installed_app_action Resource - simplemdm"
subcategory: ""
description: |-
  Triggers a one-shot action against an installed app on a device.
---

# simplemdm_installed_app_action (Resource)

Triggers a one-shot action against an installed app on a device. Supported actions:

- `update` — push an update notification.
- `request_management` — request MDM management of an unmanaged app.
- `uninstall` — DELETE the install record (uninstalls the app).

## Example Usage

```terraform
resource "simplemdm_installed_app_action" "force_update" {
  installed_app_id = "6632"
  action           = "update"
}
```

## Schema

### Required

- `installed_app_id` (String)
- `action` (String) `update`, `request_management`, or `uninstall`.

### Read-Only

- `id` (String)
- `last_triggered` (String)
