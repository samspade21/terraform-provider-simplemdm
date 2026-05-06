package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDepDevicesDataSource is opportunistic: it picks the first DEP server
// from the tenant (if any) and lists devices under it. Skips when no DEP
// server is configured. Only SIMPLEMDM_APIKEY is required.
func TestAccDepDevicesDataSource(t *testing.T) {
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

	depServerID := fmt.Sprintf("%d", servers[0].ID)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
data "simplemdm_dep_devices" "test" {
  dep_server_id = %q
}
`, depServerID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.simplemdm_dep_devices.test", "dep_server_id", depServerID),
				),
			},
		},
	})
}
