package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDeviceCustomAttributeValuesResource needs an existing device id and
// an existing custom attribute. Skips unless SIMPLEMDM_DEVICE_ID is set.
func TestAccDeviceCustomAttributeValuesResource(t *testing.T) {
	testAccPreCheck(t)
	deviceID := findFirstDeviceID(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAttributeDestroy,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "simplemdm_attribute" "fixture_a" {
  name          = "tf_acc_dev_multi_a"
  default_value = "default-a"
}

resource "simplemdm_attribute" "fixture_b" {
  name          = "tf_acc_dev_multi_b"
  default_value = "default-b"
}

resource "simplemdm_device_custom_attribute_values" "test" {
  device_id = %q

  assignments {
    name  = simplemdm_attribute.fixture_a.name
    value = "v1"
  }

  assignments {
    name  = simplemdm_attribute.fixture_b.name
    value = "v2"
  }
}
`, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_device_custom_attribute_values.test", "device_id", deviceID),
					resource.TestCheckResourceAttrSet("simplemdm_device_custom_attribute_values.test", "last_applied"),
				),
			},
		},
	})
}
