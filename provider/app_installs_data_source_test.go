package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccAppInstallsDataSource self-provisions an app, then reads its install
// records (which will likely be empty for a fresh app, but proves the read
// path works end to end). Only SIMPLEMDM_APIKEY is required.
func TestAccAppInstallsDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "simplemdm_app" "fixture" {
  app_store_id = "284882215"
  deploy_to    = "none"
}

data "simplemdm_app_installs" "test" {
  app_id = simplemdm_app.fixture.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.simplemdm_app_installs.test", "app_id",
						"simplemdm_app.fixture", "id",
					),
				),
			},
		},
	})
}
