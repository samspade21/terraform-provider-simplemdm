---
page_title: "simplemdm_munki_pkginfo Resource - simplemdm"
subcategory: ""
description: |-
  Manages the Munki pkginfo XML/PLIST blob attached to a Munki app.
---

# simplemdm_munki_pkginfo (Resource)

Manages the Munki pkginfo XML/PLIST blob attached to a Munki app. Apply uploads the supplied file via POST; Delete clears it via DELETE.

The SimpleMDM API does not expose a GET endpoint for the pkginfo content, so drift detection relies on a SHA-256 fingerprint of the locally-supplied content.

## Example Usage

```terraform
resource "simplemdm_munki_pkginfo" "internal_app" {
  app_id   = simplemdm_app.internal_app.id
  filename = "internal_app.plist"
  pkginfo  = file("${path.module}/internal_app.plist")
}
```

## Schema

### Required

- `app_id` (String) App ID to attach the pkginfo to.
- `pkginfo` (String) Pkginfo XML/PLIST contents.

### Optional

- `filename` (String) File name to send (default `munki_pkginfo.plist`).

### Read-Only

- `id` (String)
- `pkginfo_sha256` (String)
