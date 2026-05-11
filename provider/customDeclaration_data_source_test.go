package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCustomDeclarationDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCustomDeclarationDestroy,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "simplemdm_customdeclaration" "test" {
  name                      = "TF Acc Custom Declaration DS"
  declaration_type          = "com.apple.configuration.passcode.settings"
  user_scope                = false
  attribute_support         = false
  escape_attributes         = false
  reinstall_after_os_update = false
  payload                   = %s
}

data "simplemdm_customdeclaration" "test" {
  id = simplemdm_customdeclaration.test.id
}
`, "<<EOT\n"+customDeclarationTestPayload+"\nEOT"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.simplemdm_customdeclaration.test", "name", "simplemdm_customdeclaration.test", "name"),
					resource.TestCheckResourceAttrPair("data.simplemdm_customdeclaration.test", "declaration_type", "simplemdm_customdeclaration.test", "declaration_type"),
					resource.TestCheckResourceAttrPair("data.simplemdm_customdeclaration.test", "user_scope", "simplemdm_customdeclaration.test", "user_scope"),
					resource.TestCheckResourceAttrSet("data.simplemdm_customdeclaration.test", "payload"),
				),
			},
		},
	})
}
