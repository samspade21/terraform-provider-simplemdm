package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccAssignmentGroupCustomAttributeValuesDataSource self-provisions an
// assignment group and reads the custom attribute values endpoint against it
// (which will be an empty list for a fresh group). Only SIMPLEMDM_APIKEY is
// required.
func TestAccAssignmentGroupCustomAttributeValuesDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAssignmentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "simplemdm_assignmentgroup" "fixture" {
  name        = "tf-acc-cav-test"
  auto_deploy = false
}

data "simplemdm_assignmentgroup_custom_attribute_values" "test" {
  assignment_group_id = simplemdm_assignmentgroup.fixture.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.simplemdm_assignmentgroup_custom_attribute_values.test", "assignment_group_id",
						"simplemdm_assignmentgroup.fixture", "id",
					),
				),
			},
		},
	})
}
