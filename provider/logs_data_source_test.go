package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLogsDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLogsDataSourceConfig(),
				// Just verify the data source reads without error
				Check: resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

func testAccLogsDataSourceConfig() string {
	return `
data "simplemdm_logs" "test" {}
`
}
