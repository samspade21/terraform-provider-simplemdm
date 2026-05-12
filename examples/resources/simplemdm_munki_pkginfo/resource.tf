resource "simplemdm_munki_pkginfo" "internal_app" {
  app_id   = simplemdm_app.internal_app.id
  filename = "internal_app.plist"
  pkginfo  = file("${path.module}/internal_app.plist")
}
