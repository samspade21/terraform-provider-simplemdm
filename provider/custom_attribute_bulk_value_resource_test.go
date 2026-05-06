package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccCustomAttributeBulkValueResource needs an existing device id so it
// can call the bulk-set API with a real target. Skips if not supplied.
func TestAccCustomAttributeBulkValueResource(t *testing.T) {
	testAccPreCheck(t)
	deviceID := testAccRequireEnv(t, "SIMPLEMDM_DEVICE_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAttributeDestroy,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "simplemdm_attribute" "fixture" {
  name          = "tf_acc_bulk_cav"
  default_value = "default"
}

resource "simplemdm_custom_attribute_bulk_value" "test" {
  attribute_name = simplemdm_attribute.fixture.name

  assignments {
    device_id = %q
    value     = "alpha"
  }

  triggers = {
    nonce = "1"
  }
}
`, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_custom_attribute_bulk_value.test", "attribute_name", "tf_acc_bulk_cav"),
					resource.TestCheckResourceAttrSet("simplemdm_custom_attribute_bulk_value.test", "last_applied"),
				),
			},
		},
	})
}
