package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccPushCertificateSCSRDataSource verifies the SCSR endpoint reads
// without error. Only SIMPLEMDM_APIKEY is required.
func TestAccPushCertificateSCSRDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "simplemdm_push_certificate_scsr" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.simplemdm_push_certificate_scsr.test", "id", "push_certificate_scsr"),
					resource.TestCheckResourceAttrSet("data.simplemdm_push_certificate_scsr.test", "data"),
				),
			},
		},
	})
}
