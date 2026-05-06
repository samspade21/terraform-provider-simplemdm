---
page_title: "simplemdm_assignmentgroup_custom_attribute_values Data Source - simplemdm"
subcategory: ""
description: |-
  Lists custom attribute values assigned to an assignment group.
---

# simplemdm_assignmentgroup_custom_attribute_values (Data Source)

Lists custom attribute values assigned to an assignment group.

## Example Usage

```terraform
data "simplemdm_assignmentgroup_custom_attribute_values" "group" {
  assignment_group_id = "1"
}
```

## Schema

### Required

- `assignment_group_id` (String) ID of the assignment group.

### Read-Only (Block list)

Each `custom_attribute_values` block exposes `id`, `value`, `secret`, `source`.
