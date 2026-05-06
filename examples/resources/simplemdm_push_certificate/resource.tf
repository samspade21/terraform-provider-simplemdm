resource "simplemdm_push_certificate" "apns" {
  certificate = file("${path.module}/MDM_apns.pem")
  apple_id    = "ops@example.com"
}
