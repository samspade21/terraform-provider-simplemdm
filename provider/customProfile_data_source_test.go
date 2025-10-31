package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCustomProfileDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create custom profile and data source in same step
			{
				Config: providerConfig + `
					resource "simplemdm_customprofile" "test" {
						name            = "Test Custom Profile Data Source"
						mobileconfig    = file("./testfiles/testprofile.mobileconfig")
						userscope       = true
						attributesupport = false
					}

					data "simplemdm_customprofile" "test" {
						id = simplemdm_customprofile.test.id
						
						depends_on = [simplemdm_customprofile.test]
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source reads the resource correctly
					resource.TestCheckResourceAttrPair(
						"data.simplemdm_customprofile.test", "id",
						"simplemdm_customprofile.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.simplemdm_customprofile.test", "name",
						"simplemdm_customprofile.test", "name",
					),
					resource.TestCheckResourceAttrSet("data.simplemdm_customprofile.test", "mobileconfig"),
					resource.TestCheckResourceAttrSet("data.simplemdm_customprofile.test", "profileidentifier"),
				),
			},
		},
	})
}
