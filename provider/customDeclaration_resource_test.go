package provider

import (
	"context"
	"fmt"
	"testing"

	simplemdm "github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccCheckCustomDeclarationDestroy verifies the declaration was deleted
// upstream. SimpleMDM does not expose GET /custom_declarations/{id} (it 404s
// even when the record exists), so we list and look for the id.
func testAccCheckCustomDeclarationDestroy(s *terraform.State) error {
	return testAccCheckResourceDestroyed("simplemdm_customdeclaration", func(client *simplemdm.Client, id string) error {
		item, err := findCustomDeclarationByID(context.Background(), client, id)
		if err != nil {
			return err
		}
		if item == nil {
			// Mimic the "not found" sentinel that testAccCheckResourceDestroyed
			// recognises so the destroy check passes.
			return fmt.Errorf("not found")
		}
		return nil
	})(s)
}

// Uses the passcode-policy declaration type, which is universally supported on
// SimpleMDM tenants (no extra DDM-enablement gating).
const customDeclarationTestPayload = `{
  "RequirePasscode": true,
  "AllowSimple": false,
  "MinimumLength": 8
}`

const customDeclarationTestPayloadUpdated = `{
  "RequirePasscode": true,
  "AllowSimple": false,
  "MinimumLength": 10,
  "MinimumComplexCharacters": 1
}`

func TestAccCustomDeclarationResource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCustomDeclarationDestroy,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "simplemdm_customdeclaration" "test" {
  name                      = "TF Acc Custom Declaration"
  declaration_type          = "com.apple.configuration.passcode.settings"
  user_scope                = false
  attribute_support         = false
  escape_attributes         = false
  reinstall_after_os_update = false
  payload                   = %s
}
`, "<<EOT\n"+customDeclarationTestPayload+"\nEOT"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("simplemdm_customdeclaration.test", "id"),
					resource.TestCheckResourceAttr("simplemdm_customdeclaration.test", "name", "TF Acc Custom Declaration"),
					resource.TestCheckResourceAttr("simplemdm_customdeclaration.test", "declaration_type", "com.apple.configuration.passcode.settings"),
					resource.TestCheckResourceAttr("simplemdm_customdeclaration.test", "user_scope", "false"),
					resource.TestCheckResourceAttrSet("simplemdm_customdeclaration.test", "payload"),
					resource.TestCheckResourceAttrSet("simplemdm_customdeclaration.test", "profile_identifier"),
				),
			},
			{
				ResourceName:      "simplemdm_customdeclaration.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Payload is preserved from plan / state, never round-tripped
				// from /download (which adds SimpleMDM-injected fields). On
				// import there is no prior state, so the imported payload
				// will reflect the cleaned download — accept that difference.
				ImportStateVerifyIgnore: []string{"payload"},
			},
			{
				Config: providerConfig + fmt.Sprintf(`
resource "simplemdm_customdeclaration" "test" {
  name                      = "TF Acc Custom Declaration Updated"
  declaration_type          = "com.apple.configuration.passcode.settings"
  user_scope                = true
  attribute_support         = false
  escape_attributes         = false
  reinstall_after_os_update = false
  payload                   = %s
}
`, "<<EOT\n"+customDeclarationTestPayloadUpdated+"\nEOT"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_customdeclaration.test", "name", "TF Acc Custom Declaration Updated"),
					resource.TestCheckResourceAttr("simplemdm_customdeclaration.test", "user_scope", "true"),
				),
			},
		},
	})
}
