data "simplemdm_dep_device" "specific" {
  dep_server_id = "1"
  id            = "42"
}

output "model" {
  value = data.simplemdm_dep_device.specific.model
}
