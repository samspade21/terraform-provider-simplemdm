package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDepDeviceDataSource is opportunistic: it picks the first DEP server
// and the first DEP device under it from the live tenant, then re-queries the
// device through the singular data source.
func TestAccDepDeviceDataSource(t *testing.T) {
	testAccPreCheck(t)

	client, err := getTestClient()
	if err != nil {
		t.Fatalf("failed to construct test client: %v", err)
	}

	servers, err := simplemdmext.ListDepServers(context.Background(), client)
	if err != nil {
		t.Fatalf("failed to list DEP servers: %v", err)
	}
	if len(servers) == 0 {
		t.Skip("tenant has no DEP servers configured; skipping")
	}
	serverID := fmt.Sprintf("%d", servers[0].ID)

	devices, err := simplemdmext.ListDepDevices(context.Background(), client, serverID)
	if err != nil {
		t.Fatalf("failed to list DEP devices: %v", err)
	}
	if len(devices) == 0 {
		t.Skipf("DEP server %s has no devices; skipping", serverID)
	}
	deviceID := fmt.Sprintf("%d", devices[0].ID)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
data "simplemdm_dep_device" "test" {
  dep_server_id = %q
  id            = %q
}
`, serverID, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.simplemdm_dep_device.test", "id", deviceID),
				),
			},
		},
	})
}
