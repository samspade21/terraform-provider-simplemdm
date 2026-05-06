data "simplemdm_installed_app" "specific" {
  id = "6632"
}

output "version" {
  value = data.simplemdm_installed_app.specific.version
}
