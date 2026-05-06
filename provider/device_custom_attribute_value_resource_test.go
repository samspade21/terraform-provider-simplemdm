package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDeviceCustomAttributeValueResource needs an existing device id since
// SimpleMDM cannot enroll a device through the API. Skips if no fixture id is
// supplied via SIMPLEMDM_DEVICE_ID.
func TestAccDeviceCustomAttributeValueResource(t *testing.T) {
	testAccPreCheck(t)
	deviceID := testAccRequireEnv(t, "SIMPLEMDM_DEVICE_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAttributeDestroy,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "simplemdm_attribute" "fixture" {
  name          = "tf_acc_dev_cav"
  default_value = "default"
}

resource "simplemdm_device_custom_attribute_value" "test" {
  device_id      = %q
  attribute_name = simplemdm_attribute.fixture.name
  value          = "v1"
}
`, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_device_custom_attribute_value.test", "value", "v1"),
				),
			},
			{
				Config: providerConfig + fmt.Sprintf(`
resource "simplemdm_attribute" "fixture" {
  name          = "tf_acc_dev_cav"
  default_value = "default"
}

resource "simplemdm_device_custom_attribute_value" "test" {
  device_id      = %q
  attribute_name = simplemdm_attribute.fixture.name
  value          = "v2"
}
`, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_device_custom_attribute_value.test", "value", "v2"),
				),
			},
		},
	})
}
