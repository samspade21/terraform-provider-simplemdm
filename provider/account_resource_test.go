package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccAccountResource exercises the account resource using a no-op apply:
// it reads the tenant's current name + country code first, then applies them
// back unchanged. This keeps the test idempotent and non-destructive against
// the real SimpleMDM tenant. Only SIMPLEMDM_APIKEY is required.
func TestAccAccountResource(t *testing.T) {
	testAccPreCheck(t)

	client, err := getTestClient()
	if err != nil {
		t.Fatalf("failed to construct test client: %v", err)
	}

	current, err := simplemdmext.GetAccount(context.Background(), client)
	if err != nil {
		t.Fatalf("failed to read current account state: %v", err)
	}

	originalName := current.Data.Attributes.Name
	originalCC := current.Data.Attributes.AppleStoreCountryCode
	if originalCC == "" {
		// PATCH endpoint refuses an empty country code; fall back to a sane default
		// for the apply, then we keep whatever the API normalizes it to.
		originalCC = "US"
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "simplemdm_account" "tenant" {
  name                     = %q
  apple_store_country_code = %q
}
`, originalName, originalCC),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("simplemdm_account.tenant", "id"),
					resource.TestCheckResourceAttr("simplemdm_account.tenant", "name", originalName),
					resource.TestCheckResourceAttr("simplemdm_account.tenant", "apple_store_country_code", originalCC),
				),
			},
			{
				ResourceName:      "simplemdm_account.tenant",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
