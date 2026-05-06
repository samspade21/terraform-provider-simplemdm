package provider

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDeviceGroupCustomAttributeValuesDataSource opportunistically picks an
// existing device group from the tenant (device groups cannot be created via
// API). Skips if the tenant has no device groups.
func TestAccDeviceGroupCustomAttributeValuesDataSource(t *testing.T) {
	testAccPreCheck(t)

	client, err := getTestClient()
	if err != nil {
		t.Fatalf("failed to construct test client: %v", err)
	}

	groups, err := fetchAllDeviceGroups(context.Background(), client)
	if err != nil {
		t.Fatalf("failed to list device groups: %v", err)
	}
	if len(groups) == 0 {
		t.Skip("tenant has no device groups; skipping")
	}
	groupID := strconv.Itoa(groups[0].ID)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
data "simplemdm_devicegroup_custom_attribute_values" "test" {
  device_group_id = %q
}
`, groupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.simplemdm_devicegroup_custom_attribute_values.test", "device_group_id", groupID),
				),
			},
		},
	})
}
