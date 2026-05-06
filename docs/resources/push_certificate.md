---
page_title: "simplemdm_push_certificate Resource - simplemdm"
subcategory: ""
description: |-
  Manages the SimpleMDM tenant's Apple Push Notification certificate.
---

# simplemdm_push_certificate (Resource)

Manages the SimpleMDM tenant's Apple Push Notification certificate. Apply uploads the supplied PEM bytes via PUT /push_certificate. Deleting the resource only removes it from Terraform state; it does NOT clear the certificate on the SimpleMDM side.

## Example Usage

```terraform
resource "simplemdm_push_certificate" "apns" {
  certificate = file("${path.module}/MDM_apns.pem")
  apple_id    = "ops@example.com"
}
```

## Schema

### Required

- `certificate` (String, Sensitive) Push certificate PEM bytes (string), e.g. via `file("./apns.pem")`.

### Optional

- `apple_id` (String) Email address of the Apple ID the certificate was generated with.

### Read-Only

- `id` (String)
- `certificate_sha256` (String) SHA-256 fingerprint, used to detect drift.
- `expires_at` (String)
