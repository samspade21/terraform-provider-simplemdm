package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDeviceCustomAttributeValuesDataSource needs an existing device id since
// SimpleMDM cannot enroll a device through the API. Skips if no fixture id is
// supplied via SIMPLEMDM_DEVICE_ID.
func TestAccDeviceCustomAttributeValuesDataSource(t *testing.T) {
	testAccPreCheck(t)
	deviceID := findFirstDeviceID(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
data "simplemdm_device_custom_attribute_values" "test" {
  device_id = %q
}
`, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.simplemdm_device_custom_attribute_values.test", "device_id", deviceID),
				),
			},
		},
	})
}
