package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAppDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Self-provision an App Store app and read it back through the data source.
			// 284882215 is Facebook on the App Store - a long-lived ID we already use
			// elsewhere as a stable test fixture.
			{
				Config: providerConfig + `
resource "simplemdm_app" "test" {
  app_store_id = "284882215"
}

data "simplemdm_app" "test" {
  id = simplemdm_app.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.simplemdm_app.test", "id", "simplemdm_app.test", "id"),
					resource.TestCheckResourceAttrSet("data.simplemdm_app.test", "name"),
					resource.TestCheckResourceAttrSet("data.simplemdm_app.test", "bundle_id"),
					resource.TestCheckResourceAttrSet("data.simplemdm_app.test", "app_type"),
					resource.TestCheckResourceAttrSet("data.simplemdm_app.test", "platform_support"),
					resource.TestCheckResourceAttrSet("data.simplemdm_app.test", "processing_status"),
					resource.TestCheckResourceAttr("data.simplemdm_app.test", "installation_channels.#", "1"),
				),
			},
		},
	})
}
