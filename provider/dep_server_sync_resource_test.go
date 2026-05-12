package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDepServerSyncResource opportunistically triggers a sync against the
// tenant's first DEP server. Skips if no DEP server is configured.
func TestAccDepServerSyncResource(t *testing.T) {
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

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "simplemdm_dep_server_sync" "trigger" {
  dep_server_id = %q
  triggers = {
    nonce = "test-acc"
  }
}
`, serverID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_dep_server_sync.trigger", "dep_server_id", serverID),
					resource.TestCheckResourceAttrSet("simplemdm_dep_server_sync.trigger", "id"),
					resource.TestCheckResourceAttrSet("simplemdm_dep_server_sync.trigger", "last_triggered"),
				),
			},
		},
	})
}
