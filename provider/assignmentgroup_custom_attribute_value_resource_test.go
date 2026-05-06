package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccAssignmentGroupCustomAttributeValueResource self-provisions an
// assignment group and a custom attribute, then sets and updates the value.
// Only SIMPLEMDM_APIKEY is required.
func TestAccAssignmentGroupCustomAttributeValueResource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAssignmentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "simplemdm_attribute" "fixture" {
  name          = "tf_acc_ag_cav"
  default_value = "default"
}

resource "simplemdm_assignmentgroup" "fixture" {
  name        = "tf-acc-ag-cav"
  auto_deploy = false
  group_type  = "standard"
}

resource "simplemdm_assignmentgroup_custom_attribute_value" "test" {
  assignment_group_id = simplemdm_assignmentgroup.fixture.id
  attribute_name      = simplemdm_attribute.fixture.name
  value               = "alpha"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup_custom_attribute_value.test", "value", "alpha"),
				),
			},
			{
				Config: providerConfig + `
resource "simplemdm_attribute" "fixture" {
  name          = "tf_acc_ag_cav"
  default_value = "default"
}

resource "simplemdm_assignmentgroup" "fixture" {
  name        = "tf-acc-ag-cav"
  auto_deploy = false
  group_type  = "standard"
}

resource "simplemdm_assignmentgroup_custom_attribute_value" "test" {
  assignment_group_id = simplemdm_assignmentgroup.fixture.id
  attribute_name      = simplemdm_attribute.fixture.name
  value               = "bravo"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup_custom_attribute_value.test", "value", "bravo"),
				),
			},
		},
	})
}
