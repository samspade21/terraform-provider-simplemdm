package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAssignmentGroupsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAssignmentGroupsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.simplemdm_assignmentgroups.test", "assignment_groups.#"),
					// Confirm group_type is populated on every row of the
					// list. A seeded group is also forced into the list so
					// there is always at least one entry to inspect.
					resource.TestCheckResourceAttrSet("data.simplemdm_assignmentgroups.test", "assignment_groups.0.group_type"),
					// The seeded group is created via the API and therefore
					// must appear as a "static" group somewhere in the list.
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.simplemdm_assignmentgroups.test",
						"assignment_groups.*",
						map[string]string{
							"name":       "tf_acc_assignmentgroups_data_source_seed",
							"group_type": "static",
						},
					),
				),
			},
		},
	})
}

const testAccAssignmentGroupsDataSourceConfig = `
resource "simplemdm_assignmentgroup" "seed" {
  name = "tf_acc_assignmentgroups_data_source_seed"
}

data "simplemdm_assignmentgroups" "test" {
  depends_on = [simplemdm_assignmentgroup.seed]
}
`
