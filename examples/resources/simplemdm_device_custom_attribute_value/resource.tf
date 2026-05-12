resource "simplemdm_attribute" "department" {
  name          = "department"
  default_value = "engineering"
}

resource "simplemdm_device_custom_attribute_value" "alice" {
  device_id      = "1234"
  attribute_name = simplemdm_attribute.department.name
  value          = "platform-team"
}
