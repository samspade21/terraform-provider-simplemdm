package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccEnrollmentDataSource needs an enrollment-eligible device group;
// auto-discovered legacy groups don't satisfy the /enrollments endpoint,
// which 404s. Gate on SIMPLEMDM_DEVICE_GROUP_ID and skip otherwise.
func TestAccEnrollmentDataSource(t *testing.T) {
	testAccPreCheck(t)

	deviceGroupID := testAccRequireEnv(t, "SIMPLEMDM_DEVICE_GROUP_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "simplemdm_enrollment" "test" {
  device_group_id = "` + deviceGroupID + `"
}

data "simplemdm_enrollment" "test" {
  id = simplemdm_enrollment.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.simplemdm_enrollment.test", "id", "simplemdm_enrollment.test", "id"),
					resource.TestCheckResourceAttrSet("data.simplemdm_enrollment.test", "user_enrollment"),
					resource.TestCheckResourceAttrSet("data.simplemdm_enrollment.test", "welcome_screen"),
					resource.TestCheckResourceAttrSet("data.simplemdm_enrollment.test", "authentication"),
				),
			},
		},
	})
}
