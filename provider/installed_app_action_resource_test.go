package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccInstalledAppActionResource exercises the safest of the three actions
// ('update'). Skips unless SIMPLEMDM_INSTALLED_APP_ID is set, since the API
// cannot synthesize installed-app records.
func TestAccInstalledAppActionResource(t *testing.T) {
	testAccPreCheck(t)
	installedAppID := findFirstInstalledAppID(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "simplemdm_installed_app_action" "test" {
  installed_app_id = %q
  action           = "update"
}
`, installedAppID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_installed_app_action.test", "action", "update"),
					resource.TestCheckResourceAttrSet("simplemdm_installed_app_action.test", "id"),
					resource.TestCheckResourceAttrSet("simplemdm_installed_app_action.test", "last_triggered"),
				),
			},
		},
	})
}
