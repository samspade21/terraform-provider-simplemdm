package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const sampleMunkiPkgInfo = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>name</key>
  <string>tf-acc-munki</string>
  <key>display_name</key>
  <string>tf-acc-munki</string>
  <key>version</key>
  <string>1.0.0</string>
  <key>installer_type</key>
  <string>nopkg</string>
</dict>
</plist>
`

// TestAccMunkiPkgInfoResource needs an existing app id that supports munki
// pkginfo (typically a custom/enterprise app, not an App Store app). Skips
// unless SIMPLEMDM_MUNKI_APP_ID is set.
func TestAccMunkiPkgInfoResource(t *testing.T) {
	testAccPreCheck(t)
	appID := testAccRequireEnv(t, "SIMPLEMDM_MUNKI_APP_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "simplemdm_munki_pkginfo" "test" {
  app_id  = %q
  pkginfo = <<-EOT
%s
EOT
}
`, appID, sampleMunkiPkgInfo),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_munki_pkginfo.test", "app_id", appID),
					resource.TestCheckResourceAttrSet("simplemdm_munki_pkginfo.test", "pkginfo_sha256"),
				),
			},
		},
	})
}
