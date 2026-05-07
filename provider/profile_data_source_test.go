package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccProfileDataSource self-provisions a custom configuration profile and
// reads it back through the generic /profiles endpoint, which lists both
// SimpleMDM-managed and custom configuration profiles.
func TestAccProfileDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "simplemdm_customprofile" "test" {
  name         = "tf_acc_profile_data_source"
  mobileconfig = file("./testfiles/testprofile.mobileconfig")
}

data "simplemdm_profile" "test" {
  id = simplemdm_customprofile.test.id
}
`,
				// SimpleMDM exhibits eventual consistency on custom profile reads
				// shortly after creation; allow the refresh plan to differ.
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.simplemdm_profile.test", "id", "simplemdm_customprofile.test", "id"),
					resource.TestCheckResourceAttrSet("data.simplemdm_profile.test", "name"),
					resource.TestCheckResourceAttrSet("data.simplemdm_profile.test", "type"),
					resource.TestCheckResourceAttrSet("data.simplemdm_profile.test", "profile_identifier"),
					resource.TestCheckResourceAttrSet("data.simplemdm_profile.test", "user_scope"),
					resource.TestCheckResourceAttrSet("data.simplemdm_profile.test", "reinstall_after_os_update"),
					resource.TestCheckResourceAttrSet("data.simplemdm_profile.test", "group_count"),
					resource.TestCheckResourceAttrSet("data.simplemdm_profile.test", "device_count"),
				),
			},
		},
	})
}
