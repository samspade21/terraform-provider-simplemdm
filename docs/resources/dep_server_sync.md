---
page_title: "simplemdm_dep_server_sync Resource - simplemdm"
subcategory: ""
description: |-
  Triggers a sync of the given DEP server with Apple Business Manager.
---

# simplemdm_dep_server_sync (Resource)

Triggers a sync of the given DEP server with Apple Business Manager. This is a fire-and-forget action: the sync runs once on Create. Change any key in the `triggers` map (or remove and re-add the resource) to re-run the sync.

## Example Usage

```terraform
resource "simplemdm_dep_server_sync" "manual" {
  dep_server_id = "1"
  triggers = {
    # Change any value to retrigger the sync.
    nonce = "2025-05-06"
  }
}
```

## Schema

### Required

- `dep_server_id` (String)

### Optional

- `triggers` (Map of String) Optional arbitrary string map. Changing any value forces resource replacement, re-running the sync.

### Read-Only

- `id` (String) Synthetic ID for the sync action.
- `last_triggered` (String) RFC3339 timestamp of the most recent sync invocation.
