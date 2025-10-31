---
page_title: "simplemdm_app Resource - terraform-provider-simplemdm"
subcategory: ""
description: |-
  Manages apps in SimpleMDM, including App Store apps, Volume Purchase Program (VPP) apps, and custom enterprise/B2B apps.
---

# simplemdm_app (Resource)

Manages apps in SimpleMDM. This resource allows you to add and configure apps from the Apple App Store, deploy Volume Purchase Program (VPP) apps using bundle identifiers, or upload custom enterprise and macOS package apps.

## Example Usage

### Basic Example - App Store App

```terraform
resource "simplemdm_app" "slack" {
  app_store_id = "803453959"
  name         = "Slack"
  deploy_to    = "all"
}
```

### Advanced Example - VPP App with Bundle ID

```terraform
# Deploy an existing Volume Purchase Program (VPP) app by referencing its
# bundle identifier. SimpleMDM resolves the bundle ID to the latest catalog
# metadata and makes it available to devices.
resource "simplemdm_app" "company_app" {
  bundle_id = "com.example.myapp"
  name      = "Company Internal App"
  deploy_to = "outdated"
}

# Reference computed attributes
output "app_version" {
  description = "Current version of the app"
  value       = simplemdm_app.company_app.version
}

output "app_installation_channels" {
  description = "Deployment channels supported by the app"
  value       = simplemdm_app.company_app.installation_channels
}
```

### Advanced Example - Custom Enterprise App

```terraform
# Upload a custom enterprise or macOS package app by providing a binary file.
# The provider will post the binary to SimpleMDM and keep the metadata in sync.
resource "simplemdm_app" "enterprise_tools" {
  name        = "Internal Tools Suite"
  binary_file = "${path.module}/files/internal-tools.pkg"
  deploy_to   = "none"
}

# Monitor processing status
output "processing_status" {
  description = "Processing state for the uploaded enterprise app binary"
  value       = simplemdm_app.enterprise_tools.processing_status
}
```

### Advanced Example - App with Assignment Group

```terraform
resource "simplemdm_app" "microsoft_teams" {
  app_store_id = "1113153706"
  name         = "Microsoft Teams"
  deploy_to    = "all"
}

resource "simplemdm_assignmentgroup" "sales_team" {
  name        = "Sales Team"
  auto_deploy = true
  apps        = [simplemdm_app.microsoft_teams.id]
}
```

## Schema

### Optional

- `app_store_id` (String) The Apple App Store ID of the app to be added. Example: `1090161858`. Required when adding App Store apps.
- `binary_file` (String) Absolute or relative path to an app binary (`.ipa` or `.pkg`) to upload. Required when managing enterprise, custom B2B, or macOS package apps.
- `bundle_id` (String) The bundle identifier of the Apple App Store app to be added. Example: `com.example.MyApp`. Required when deploying VPP apps by bundle ID.
- `deploy_to` (String) Deploy the app to associated devices immediately after the app has been uploaded and processed. Valid values: `none`, `outdated`, `all`. Default: `none`.
- `name` (String) The name that SimpleMDM will use to reference this app. If left blank, SimpleMDM will automatically set this to the app name specified by the binary.

### Read-Only

- `app_type` (String) The catalog classification of the app (e.g., `app store`, `enterprise`, `custom b2b`).
- `created_at` (String) Timestamp when the app was added to SimpleMDM.
- `id` (String) The unique identifier of the app in SimpleMDM.
- `installation_channels` (List of String) The deployment channels supported by the app.
- `platform_support` (String) The platform supported by the app (`iOS` or `macOS`).
- `processing_status` (String) The current processing status of the app binary within SimpleMDM.
- `status` (String) The current deployment status of the app.
- `updated_at` (String) Timestamp when the app was last updated in SimpleMDM.
- `version` (String) The latest version reported by SimpleMDM for the app.

## Import

Import is supported using the following syntax:

```shell
# App can be imported by specifying the app ID
terraform import simplemdm_app.example 123456
```

## Notes

- **App Type Selection**: You must specify exactly one of `app_store_id`, `bundle_id`, or `binary_file`.
- **Binary Upload**: When uploading custom apps via `binary_file`, monitor the `processing_status` attribute to ensure successful processing.
- **Deployment**: Setting `deploy_to` to `all` or `outdated` will automatically deploy the app to devices in assigned groups.
- **Version Management**: The `version` attribute is read-only and reflects the latest version available in SimpleMDM's catalog.