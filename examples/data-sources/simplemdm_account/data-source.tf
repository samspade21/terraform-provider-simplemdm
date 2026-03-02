data "simplemdm_account" "current" {}

output "account_name" {
  value = data.simplemdm_account.current.name
}

output "dep_enabled" {
  value = data.simplemdm_account.current.dep_enabled
}
