---
page_title: "simplemdm_log Data Source - simplemdm"
subcategory: ""
description: |-
  Retrieves a specific SimpleMDM log entry by ID.
---

# simplemdm_log (Data Source)

Retrieves a specific SimpleMDM log entry by ID.

## Example Usage

```terraform
data "simplemdm_log" "specific" {
  id = "964595dfd5be464a82cbb9019f55d82b"
}

output "log_message" {
  value = data.simplemdm_log.specific.message
}
```

## Schema

### Required

- `id` (String) The ID of the log entry to retrieve.

### Read-Only

- `namespace` (String)
- `source` (String)
- `event_type` (String)
- `level` (String)
- `message` (String)
- `at` (String) The timestamp when the log entry was created.
- `metadata` (String) Raw JSON metadata associated with the log entry.
