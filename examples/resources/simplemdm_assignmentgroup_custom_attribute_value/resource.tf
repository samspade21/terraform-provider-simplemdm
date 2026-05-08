resource "simplemdm_attribute" "region" {
  name          = "region"
  default_value = "us"
}

resource "simplemdm_assignmentgroup" "engineering" {
  name        = "engineering"
  auto_deploy = true
}

resource "simplemdm_assignmentgroup_custom_attribute_value" "engineering_region" {
  assignment_group_id = simplemdm_assignmentgroup.engineering.id
  attribute_name      = simplemdm_attribute.region.name
  value               = "eu"
}
