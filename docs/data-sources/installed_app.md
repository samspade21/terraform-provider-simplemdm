---
page_title: "simplemdm_installed_app Data Source - simplemdm"
subcategory: ""
description: |-
  Retrieves a single installed app record by ID.
---

# simplemdm_installed_app (Data Source)

Retrieves a single installed app record by ID.

## Example Usage

```terraform
data "simplemdm_installed_app" "specific" {
  id = "6632"
}
```

## Schema

### Required

- `id` (String) Installed app ID.

### Read-Only

- `name`, `identifier`, `version`, `short_version` (String)
- `bundle_size`, `dynamic_size` (Number, bytes)
- `managed` (Bool)
- `discovered_at`, `last_seen_at` (String)
