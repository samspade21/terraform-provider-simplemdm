package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeviceGroupDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create device group first
			{
				Config: providerConfig + `
					resource "simplemdm_devicegroup" "test" {
						name = "Test Data Source Device Group"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("simplemdm_devicegroup.test", "id"),
					resource.TestCheckResourceAttr("simplemdm_devicegroup.test", "name", "Test Data Source Device Group"),
				),
			},
			// Then test reading it with data source
			{
				Config: providerConfig + `
					resource "simplemdm_devicegroup" "test" {
						name = "Test Data Source Device Group"
					}

					data "simplemdm_devicegroup" "test" {
						id = simplemdm_devicegroup.test.id
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source matches resource
					resource.TestCheckResourceAttrPair(
						"data.simplemdm_devicegroup.test", "id",
						"simplemdm_devicegroup.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.simplemdm_devicegroup.test", "name",
						"simplemdm_devicegroup.test", "name",
					),
				),
			},
		},
	})
}
