# Terraform Provider for SimpleMDM

This repository contains the Terraform provider that manages resources in
[SimpleMDM](https://simplemdm.com). The provider is implemented with the
[Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework)
and includes its own SimpleMDM API client under
[`internal/simplemdm/`](./internal/simplemdm/). The client was originally a
separate Go module (`github.com/DavidKrau/simplemdm-go-client`) that the
provider depended on externally; it has been vendored in so we can fix bugs
in place and add coverage for endpoints the upstream module didn't expose.

> **Compatibility:** This provider targets the **SimpleMDM API v1.55**
> ([api.simplemdm.com/v1](https://api.simplemdm.com/v1)). Endpoints and
> response shapes added or changed after that version are not guaranteed to
> work; check the API changelog if you hit unexpected behaviour.

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

Most acceptance tests self-provision their own fixtures (apps, attributes,
custom profiles, scripts, assignment groups, enrollments, etc.) directly via
the provider, so the bare minimum above is enough to exercise the bulk of the
suite. Tests that touch real-world hardware fall back to API-side discovery —
they query the tenant for an existing device, device group, installed app, or
DEP server and skip cleanly when none is present. GitHub Actions runs the same
command in [`.github/workflows/test.yml`](.github/workflows/test.yml).

#### Always required

| Variable | Purpose |
|----------|---------|
| `TF_ACC` | Set to `1` to opt in to acceptance tests at all. Without it every test skips. |
| `SIMPLEMDM_APIKEY` | API key for the test tenant. Without it every acceptance test skips. |
| `SIMPLEMDM_HOST` | Optional. Defaults to `a.simplemdm.com`. |

#### Auto-discovered (no env var needed; tests skip cleanly if not present)

The following resources are looked up at runtime via the SimpleMDM API. Each
test honours an env-var override (column 3) if you want to pin a specific
fixture, but normally you can leave them unset.

| Resource | Lookup endpoint | Override |
|----------|-----------------|----------|
| First enrolled device | `GET /devices` | `SIMPLEMDM_DEVICE_ID` |
| First device group | `GET /device_groups` | `SIMPLEMDM_DEVICE_GROUP_ID` |
| First installed-app record on any device | `GET /devices/{id}/installed_apps` | `SIMPLEMDM_INSTALLED_APP_ID` |
| First non-App-Store app (enterprise / custom B2B) | `GET /apps` filtered by `app_type` | `SIMPLEMDM_MUNKI_APP_ID` |
| First DEP server | `GET /dep_servers` | `SIMPLEMDM_DEP_SERVER_ID` |

Tests that rely on auto-discovery and find nothing print a message like
`Acceptance test skipped: an enrolled device not found in tenant; set
SIMPLEMDM_DEVICE_ID to override.` and exit cleanly.

#### Manual override required (specific tenant fixtures)

A few tests need a specific fixture that auto-discovery can't reliably pick:

| Variable | Tests | Why a manual fixture is needed |
|----------|-------|-------------------------------|
| `SIMPLEMDM_DEVICE_GROUP_ID` | `simplemdm_enrollment` resource + data source, `simplemdm_scriptjob` resource | The `/enrollments` and `/script_jobs` endpoints reject newly-created assignment groups *and* arbitrary auto-discovered legacy device groups — they need a specific group registered as enrollment-eligible by SimpleMDM. |
| `SIMPLEMDM_PUSH_CERT_PEM` | `simplemdm_push_certificate` resource | Only Apple can issue a real APNs PEM; there's nothing to discover. |
| `SIMPLEMDM_PUSH_CERT_APPLE_ID` | `simplemdm_push_certificate` resource | Pairs with `SIMPLEMDM_PUSH_CERT_PEM`; the Apple ID the cert was generated under. |

## Migration notes

* **`simplemdm_devicegroup` resource and data sources have been removed.**
  The SimpleMDM API marked `/device_groups` deprecated in favour of
  `/assignment_groups`. Replace `resource "simplemdm_devicegroup" "x" { … }`
  with `resource "simplemdm_assignmentgroup" "x" { … }` and migrate any
  references. (`simplemdm_devicegroup` data source / `simplemdm_devicegroups`
  list / `simplemdm_devicegroup_custom_attribute_values` were removed in the
  same change.)
* **`simplemdm_assignmentgroup.install_type` has been removed.** The field
  was deprecated by the SimpleMDM API and would force a resource replacement
  on every plan. Set `install_type` per-app via the Assign App endpoint
  instead (or via the `apps` set with the relevant deployment).
* **`simplemdm_assignmentgroup.group_type` has been removed.** SimpleMDM
  deprecated the `type` write parameter, and on New Groups Experience accounts
  the response field returns `"static"` / `"dynamic"` rather than the
  configured `"standard"` / `"munki"` — drift was guaranteed. Use per-app
  `apps_deployment_types` (`standard`/`munki`) on the assignment group to
  control deployment type instead. The same field has been removed from the
  `simplemdm_assignmentgroup` and `simplemdm_assignmentgroups` data sources.
* **`simplemdm_profile` resource never existed in this provider** — only the
  read-only `simplemdm_profile` data source. To create profiles via Terraform
  use `simplemdm_customprofile` (custom mobileconfig) or
  `simplemdm_customdeclaration` (Declarative Device Management).

## Known issues

* Device name updates require a manual PATCH request outside of Terraform.
* Profiles and custom profiles applied to assignment groups or devices cannot be updated via API; Terraform compares the desired configuration against the previous state only.
