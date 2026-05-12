data "simplemdm_push_certificate" "current" {}

output "certificate_expires_at" {
  value = data.simplemdm_push_certificate.current.expires_at
}

output "apple_id" {
  value = data.simplemdm_push_certificate.current.apple_id
}
