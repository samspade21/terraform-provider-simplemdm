---
page_title: "simplemdm_account Resource - simplemdm"
subcategory: ""
description: |-
  Manages SimpleMDM account-level settings (singleton). Creating this resource overwrites the existing tenant settings; deleting it only removes the resource from Terraform state and does not modify the tenant.
---

# simplemdm_account (Resource)

Manages SimpleMDM account-level settings (singleton). Creating this resource overwrites the existing tenant settings; deleting it only removes the resource from Terraform state and does not modify the tenant.

There is only one account per SimpleMDM tenant. Use this resource to declare the tenant's company name and Apple App Store country code.

## Example Usage

```terraform
resource "simplemdm_account" "tenant" {
  name                     = "ACME, Inc."
  apple_store_country_code = "US"
}
```

## Schema

### Optional

- `name` (String) Company name associated with the account.
- `apple_store_country_code` (String) Apple App Store country code (e.g. US, AU).

### Read-Only

- `id` (String) Numeric SimpleMDM account ID.

## Import

The account is a singleton; import using its numeric ID:

```shell
terraform import simplemdm_account.tenant 12345
```
