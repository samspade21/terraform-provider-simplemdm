data "simplemdm_dep_devices" "all" {
  dep_server_id = "1"
}

output "serial_numbers" {
  value = [for d in data.simplemdm_dep_devices.all.dep_devices : d.serial_number]
}
