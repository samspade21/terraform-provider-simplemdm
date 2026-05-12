data "simplemdm_log" "specific" {
  id = "964595dfd5be464a82cbb9019f55d82b"
}

output "log_message" {
  value = data.simplemdm_log.specific.message
}
