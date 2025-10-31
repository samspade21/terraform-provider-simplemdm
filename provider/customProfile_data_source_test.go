package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCustomProfileDataSource(t *testing.T) {
	testAccPreCheck(t)

	// Use existing custom profile for data source test
	// Creating and immediately reading causes API timing issues
	customProfileID := testAccRequireEnv(t, "SIMPLEMDM_CUSTOM_PROFILE_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read existing custom profile
			{
				Config: providerConfig + `
					data "simplemdm_customprofile" "test" {
						id = "` + customProfileID + `"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.simplemdm_customprofile.test", "id", customProfileID),
					resource.TestCheckResourceAttrSet("data.simplemdm_customprofile.test", "name"),
					resource.TestCheckResourceAttrSet("data.simplemdm_customprofile.test", "mobileconfig"),
					resource.TestCheckResourceAttrSet("data.simplemdm_customprofile.test", "profileidentifier"),
				),
			},
		},
	})
}
