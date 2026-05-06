data "simplemdm_push_certificate_scsr" "csr" {}

output "scsr_data" {
  description = "Base64-encoded plist to upload to Apple Push Certificates Portal."
  value       = data.simplemdm_push_certificate_scsr.csr.data
  sensitive   = true
}
