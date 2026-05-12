data "simplemdm_app_installs" "example" {
  app_id = "1234"
}

output "managed_devices" {
  value = [
    for install in data.simplemdm_app_installs.example.installs :
    install.device_id if install.managed
  ]
}
