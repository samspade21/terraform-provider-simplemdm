resource "simplemdm_attribute" "office" {
  name          = "office"
  default_value = "remote"
}

resource "simplemdm_custom_attribute_bulk_value" "office_assignments" {
  attribute_name = simplemdm_attribute.office.name

  assignments {
    device_id = "1234"
    value     = "berlin"
  }

  assignments {
    device_id = "5678"
    value     = "london"
  }

  triggers = {
    nonce = "2025-05-06"
  }
}
