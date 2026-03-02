package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDepServersDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDepServersDataSourceConfig(),
				// DEP servers may or may not be configured; we just verify the data source reads without error.
				Check: resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testAccDepServersDataSourceConfig() string {
	return `
data "simplemdm_dep_servers" "test" {}
`
}
