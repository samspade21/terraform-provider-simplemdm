package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccInstalledAppDataSource needs an existing installed app id since
// SimpleMDM cannot create installed-app records via API. Skips if not set.
func TestAccInstalledAppDataSource(t *testing.T) {
	testAccPreCheck(t)
	installedAppID := testAccRequireEnv(t, "SIMPLEMDM_INSTALLED_APP_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
data "simplemdm_installed_app" "test" {
  id = %q
}
`, installedAppID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.simplemdm_installed_app.test", "id", installedAppID),
					resource.TestCheckResourceAttrSet("data.simplemdm_installed_app.test", "identifier"),
				),
			},
		},
	})
}
