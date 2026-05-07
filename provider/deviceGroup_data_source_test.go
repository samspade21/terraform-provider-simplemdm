package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDeviceGroupDataSource auto-discovers an existing device group via
// the API. Device groups can't be created via API (the endpoint is
// deprecated), so the test uses whatever the tenant already has and skips
// when the tenant has none.
func TestAccDeviceGroupDataSource(t *testing.T) {
	testAccPreCheck(t)

	deviceGroupID := findFirstDeviceGroupID(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "simplemdm_devicegroup" "test" {
  id = "` + deviceGroupID + `"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.simplemdm_devicegroup.test", "id", deviceGroupID),
					resource.TestCheckResourceAttrSet("data.simplemdm_devicegroup.test", "name"),
				),
			},
		},
	})
}
