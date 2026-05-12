package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAttributeDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "simplemdm_attribute" "test" {
  name          = "tf_acc_test_attribute"
  default_value = "test_default"
}

data "simplemdm_attribute" "test" {
  name = simplemdm_attribute.test.name
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.simplemdm_attribute.test", "name", "tf_acc_test_attribute"),
					resource.TestCheckResourceAttr("data.simplemdm_attribute.test", "default_value", "test_default"),
				),
			},
		},
	})
}
