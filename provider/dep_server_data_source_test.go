package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDepServerDataSource(t *testing.T) {
	testAccPreCheck(t)
	serverID := testAccGetEnv(t, "SIMPLEMDM_DEP_SERVER_ID")
	if serverID == "" {
		t.Skip("SIMPLEMDM_DEP_SERVER_ID not set, skipping DEP server data source test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDepServerDataSourceConfig(serverID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.simplemdm_dep_server.test", "id", serverID),
					resource.TestCheckResourceAttrSet("data.simplemdm_dep_server.test", "server_name"),
				),
			},
		},
	})
}

func testAccDepServerDataSourceConfig(serverID string) string {
	return fmt.Sprintf(`
data "simplemdm_dep_server" "test" {
  id = "%s"
}
`, serverID)
}
