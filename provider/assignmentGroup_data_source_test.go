package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAssignmentGroupDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "simplemdm_assignmentgroup" "test" {
  name = "tf_acc_assignmentgroup_data_source"
}

data "simplemdm_assignmentgroup" "test" {
  id = simplemdm_assignmentgroup.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.simplemdm_assignmentgroup.test", "id", "simplemdm_assignmentgroup.test", "id"),
					resource.TestCheckResourceAttr("data.simplemdm_assignmentgroup.test", "name", "tf_acc_assignmentgroup_data_source"),
				),
			},
		},
	})
}
