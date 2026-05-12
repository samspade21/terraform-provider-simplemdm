resource "simplemdm_device_custom_attribute_values" "alice_phone" {
  device_id = "1234"

  assignments {
    name  = "department"
    value = "platform-team"
  }

  assignments {
    name  = "office"
    value = "berlin"
  }
}
