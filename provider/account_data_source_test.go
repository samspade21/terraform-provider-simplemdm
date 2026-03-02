package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAccountDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.simplemdm_account.test", "id"),
					resource.TestCheckResourceAttrSet("data.simplemdm_account.test", "name"),
				),
			},
		},
	})
}

func testAccAccountDataSourceConfig() string {
	return `
data "simplemdm_account" "test" {}
`
}
