---
page_title: "simplemdm_assignmentgroup_custom_attribute_value Resource - simplemdm"
subcategory: ""
description: |-
  Sets a custom attribute value on an assignment group.
---

# simplemdm_assignmentgroup_custom_attribute_value (Resource)

Sets a custom attribute value on an assignment group. Deleting the resource clears the value.

## Example Usage

```terraform
resource "simplemdm_assignmentgroup_custom_attribute_value" "engineering_region" {
  assignment_group_id = simplemdm_assignmentgroup.engineering.id
  attribute_name      = simplemdm_attribute.region.name
  value               = "eu"
}
```

## Schema

### Required

- `assignment_group_id` (String)
- `attribute_name` (String)
- `value` (String)

### Read-Only

- `id` (String) Composite ID `<assignment_group_id>/<attribute_name>`.
