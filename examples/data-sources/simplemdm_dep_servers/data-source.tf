data "simplemdm_dep_servers" "all" {}

output "dep_server_count" {
  value = length(data.simplemdm_dep_servers.all.dep_servers)
}
