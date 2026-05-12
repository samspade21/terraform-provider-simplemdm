data "simplemdm_dep_server" "example" {
  id = "1"
}

output "dep_server_name" {
  value = data.simplemdm_dep_server.example.server_name
}
