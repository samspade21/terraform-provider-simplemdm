---
page_title: "simplemdm_push_certificate_scsr Data Source - simplemdm"
subcategory: ""
description: |-
  Retrieves a base64-encoded plist (signed CSR) for upload to the Apple Push Certificates Portal.
---

# simplemdm_push_certificate_scsr (Data Source)

Retrieves a base64-encoded plist (signed CSR) that you upload to the Apple Push Certificates Portal when generating or renewing the APNs certificate for your SimpleMDM tenant.

## Example Usage

```terraform
data "simplemdm_push_certificate_scsr" "csr" {}

output "scsr_data" {
  description = "Base64-encoded plist to upload to Apple Push Certificates Portal."
  value       = data.simplemdm_push_certificate_scsr.csr.data
  sensitive   = true
}
```

## Schema

### Read-Only

- `id` (String)
- `data` (String) Base64-encoded plist value to upload to Apple as-is.
