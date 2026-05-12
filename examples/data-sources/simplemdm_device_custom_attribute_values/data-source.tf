data "simplemdm_device_custom_attribute_values" "device42" {
  device_id = "42"
}

output "department" {
  value = [
    for v in data.simplemdm_device_custom_attribute_values.device42.custom_attribute_values :
    v.value if v.id == "department"
  ]
}
