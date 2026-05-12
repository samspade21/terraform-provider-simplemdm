package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccLogDataSource looks up an arbitrary log id from the tenant's log
// stream and re-queries it via the singular data source. Skips if the tenant
// has no logs.
func TestAccLogDataSource(t *testing.T) {
	testAccPreCheck(t)

	client, err := getTestClient()
	if err != nil {
		t.Fatalf("failed to construct test client: %v", err)
	}

	logs, err := simplemdmext.ListLogs(context.Background(), client)
	if err != nil {
		t.Fatalf("unable to list logs to seed test: %v", err)
	}
	if len(logs) == 0 {
		t.Skip("tenant has no log entries; skipping single-log test")
	}

	logID := logs[0].ID

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
data "simplemdm_log" "test" {
  id = %q
}
`, logID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.simplemdm_log.test", "id", logID),
					resource.TestCheckResourceAttrSet("data.simplemdm_log.test", "namespace"),
				),
			},
		},
	})
}
