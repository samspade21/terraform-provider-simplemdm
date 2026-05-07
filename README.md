# Terraform Provider for SimpleMDM

This repository contains the Terraform provider that manages resources in
[SimpleMDM](https://simplemdm.com). The provider is implemented with the
[Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework)
and uses the official
[`simplemdm-go-client`](https://github.com/DavidKrau/simplemdm-go-client) to talk to
SimpleMDM's REST API.

## Installation

Add the provider to your Terraform configuration. Published releases are available on
[the Terraform Registry](https://registry.terraform.io/providers/DavidKrau/simplemdm):

```terraform
terraform {
  required_providers {
    simplemdm = {
      source  = "DavidKrau/simplemdm"
      version = "~> 0.1"
    }
  }
}

provider "simplemdm" {
  # Optional. Defaults to https://a.simplemdm.com
  host   = "a.simplemdm.com"

  # Required unless provided via the SIMPLEMDM_APIKEY environment variable.
  apikey = var.simplemdm_api_key
}
```

The provider accepts two configuration attributes:

| Attribute | Environment variable | Notes |
|-----------|----------------------|-------|
| `apikey`  | `SIMPLEMDM_APIKEY`   | Required. API key for your tenant. |
| `host`    | `SIMPLEMDM_HOST`     | Optional. Override the API hostname (defaults to `a.simplemdm.com`). |

## Documentation and examples

* Generated documentation for every resource and data source lives in [`docs/`](./docs/).
* Copyable end-to-end examples for the provider, resources, and data sources are in [`examples/`](./examples/).

Regenerate the documentation whenever schemas change:

```bash
go generate ./...
```

## Project layout

| Path | Description |
|------|-------------|
| [`provider/`](./provider/) | Provider, resource, data source, and acceptance test implementations. |
| [`internal/`](./internal/) | Helper packages that support the provider. |
| [`docs/`](./docs/) | Terraform Plugin Docs output used by the Terraform Registry. |
| [`examples/`](./examples/) | Working configuration snippets for the provider. |
| [`scripts/`](./scripts/) | Utility scripts, including fixture discovery helpers for tests. |

## Development workflow

This project uses Go 1.24 (see [`go.mod`](./go.mod)). Typical development steps:

```bash
# Install dependencies and verify the provider builds
go mod download
go build ./...

# Run linting (same tool as CI)
golangci-lint run

# Run unit tests
go test ./...
```

### Pre-commit hooks

Formatting expectations are codified via
[`pre-commit`](https://pre-commit.com/). The configured hooks run `gofmt` and a
few general hygiene checks before each commit. Install and enable the hooks
locally with:

```bash
pip install pre-commit  # once per development environment
pre-commit install
```

Run the hooks manually across the entire repository with `pre-commit run --all-files`.

### Acceptance tests

Acceptance tests live under [`provider/`](./provider/). The default run only
needs the API credentials:

```bash
TF_ACC=1 SIMPLEMDM_APIKEY="your-api-key" go test -v -cover ./provider/
```

Tests requiring tenant-specific fixtures (devices, push certificates, etc.)
call `testAccRequireEnv` and skip cleanly when the relevant variable is unset.
GitHub Actions runs the same command in
[`.github/workflows/test.yml`](.github/workflows/test.yml).

#### Required (always)

| Variable | Purpose |
|----------|---------|
| `TF_ACC` | Set to `1` to opt in to acceptance tests at all. Without it every test skips. |
| `SIMPLEMDM_APIKEY` | API key for the test tenant. Without it every acceptance test skips. |
| `SIMPLEMDM_HOST` | Optional. Defaults to `a.simplemdm.com`. |

#### Optional fixtures

Fixture variables unlock additional acceptance tests. Each one references an
existing object in the test tenant; tests that need it are listed under
"Unlocks". Tests that don't reference a fixture self-provision their own
resources during the run.

##### Apps & assignment groups

| Variable | Unlocks | Notes |
|----------|---------|-------|
| `SIMPLEMDM_APP_ID` | `simplemdm_app` data source; `simplemdm_assignmentgroup` resource update path. | ID of any existing app. |
| `SIMPLEMDM_ASSIGNMENT_GROUP_ID` | `simplemdm_assignmentgroup` resource and data source. | ID of an assignment group (the modern replacement for device groups). |
| `SIMPLEMDM_PROFILE_ID` | `simplemdm_profile` data source; `simplemdm_assignmentgroup` resource. | ID of a profile created in the SimpleMDM UI. |

##### Custom attributes

| Variable | Unlocks | Notes |
|----------|---------|-------|
| `SIMPLEMDM_ATTRIBUTE_NAME` | `simplemdm_attribute` data source. | Name of an existing custom attribute. |

##### Custom declarations

| Variable | Unlocks | Notes |
|----------|---------|-------|
| `SIMPLEMDM_CUSTOM_DECLARATION_DEVICE_ID` | `simplemdm_customdeclaration_device_assignment` resource. | Device capable of receiving DDM declarations. |

##### Device groups (legacy)

These tests cover the deprecated `simplemdm_devicegroup` resource. New
deployments should use `simplemdm_assignmentgroup` instead.

| Variable | Unlocks | Notes |
|----------|---------|-------|
| `SIMPLEMDM_DEVICE_GROUP_ID` | `simplemdm_devicegroup` data source; `simplemdm_device`, `simplemdm_enrollment`, `simplemdm_scriptjob` resources. | Existing legacy device group ID. |
| `SIMPLEMDM_DEVICE_GROUP_CLONE_SOURCE_ID` | `simplemdm_devicegroup` resource clone path. | Source group used for the clone test. |
| `SIMPLEMDM_DEVICE_GROUP_NAME` | `simplemdm_devicegroup` resource. | Name reused across the resource's update steps. |
| `SIMPLEMDM_DEVICE_GROUP_ATTRIBUTE_KEY` | `simplemdm_devicegroup` resource. | Attribute key written during the update test. |
| `SIMPLEMDM_DEVICE_GROUP_ATTRIBUTE_VALUE` | `simplemdm_devicegroup` resource. | Initial value for the attribute key above. |
| `SIMPLEMDM_DEVICE_GROUP_ATTRIBUTE_UPDATED_VALUE` | `simplemdm_devicegroup` resource. | Updated value to verify the update path. |
| `SIMPLEMDM_DEVICE_GROUP_PROFILE_ID` | `simplemdm_device`, `simplemdm_devicegroup` resources. | Profile attached during the resource lifecycle tests. |
| `SIMPLEMDM_DEVICE_GROUP_PROFILE_UPDATED_ID` | `simplemdm_device`, `simplemdm_devicegroup` resources. | Replacement profile ID to drive the update. |
| `SIMPLEMDM_DEVICE_GROUP_CUSTOM_PROFILE_ID` | `simplemdm_device`, `simplemdm_devicegroup` resources. | Custom profile attached during lifecycle tests. |
| `SIMPLEMDM_DEVICE_GROUP_CUSTOM_PROFILE_UPDATED_ID` | `simplemdm_device`, `simplemdm_devicegroup` resources. | Replacement custom profile ID. |

##### Devices

A single enrolled device unlocks the largest block of acceptance tests:

| Variable | Unlocks | Notes |
|----------|---------|-------|
| `SIMPLEMDM_DEVICE_ID` | `simplemdm_device` data source; `simplemdm_device_profiles`, `simplemdm_device_installed_apps`, `simplemdm_device_users`, `simplemdm_device_custom_attribute_values` data sources; `simplemdm_device_command`, `simplemdm_device_custom_attribute_value`, `simplemdm_device_custom_attribute_values`, `simplemdm_custom_attribute_bulk_value` resources. | ID of an enrolled device. The device must be macOS for `simplemdm_device_users` to return data. |

##### DEP / Apple Business Manager

| Variable | Unlocks | Notes |
|----------|---------|-------|
| `SIMPLEMDM_DEP_SERVER_ID` | `simplemdm_dep_server` data source. | Optional. Auto-detected from the first DEP server in the tenant when unset. |

##### Enrollments

| Variable | Unlocks | Notes |
|----------|---------|-------|
| `SIMPLEMDM_ENROLLMENT_ID` | `simplemdm_enrollment` data source. | ID of an existing enrollment. |
| `SIMPLEMDM_ENROLLMENT_CONTACT` | `simplemdm_enrollment` resource invitation path. | Email or `+`-prefixed phone for the invitation step. |
| `SIMPLEMDM_ENROLLMENT_CONTACT_UPDATE` | `simplemdm_enrollment` resource update step. | Replacement contact value to drive the update test. |

##### Installed apps & Munki

| Variable | Unlocks | Notes |
|----------|---------|-------|
| `SIMPLEMDM_INSTALLED_APP_ID` | `simplemdm_installed_app` data source; `simplemdm_installed_app_action` resource. | Per-device installed-app record ID, *not* a catalog app ID. |
| `SIMPLEMDM_MUNKI_APP_ID` | `simplemdm_munki_pkginfo` resource. | ID of a custom (non-App-Store) app that supports Munki pkginfo. |

##### Push certificate

| Variable | Unlocks | Notes |
|----------|---------|-------|
| `SIMPLEMDM_PUSH_CERT_PEM` | `simplemdm_push_certificate` resource. | Filesystem path to a real APNs PEM file. |
| `SIMPLEMDM_PUSH_CERT_APPLE_ID` | `simplemdm_push_certificate` resource. | Optional Apple ID email used when uploading. |

##### Scripts

| Variable | Unlocks | Notes |
|----------|---------|-------|
| `SIMPLEMDM_SCRIPT_ID` | `simplemdm_script` data source. | ID of an existing script. |
| `SIMPLEMDM_SCRIPT_JOB_ID` | `simplemdm_scriptjob` data source. | ID of an existing script job (jobs disappear after one month). |

Use [`scripts/discover-test-fixtures.sh`](./scripts/discover-test-fixtures.sh)
to collect most of the IDs above automatically from your tenant and emit `gh
secret set` commands that match the CI workflow.

## Known issues

* Device groups are deprecated in SimpleMDM. The legacy `simplemdm_devicegroup` resource and data source remain for backward compatibility, but new deployments should favor `simplemdm_assignmentgroup`.
* Device name updates require a manual PATCH request outside of Terraform.
* Profiles and custom profiles applied to assignment groups or devices cannot be updated via API; Terraform compares the desired configuration against the previous state only.
