data "simplemdm_logs" "recent" {}

output "log_count" {
  value = length(data.simplemdm_logs.recent.logs)
}
