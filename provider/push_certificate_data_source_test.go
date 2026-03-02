package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPushCertificateDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPushCertificateDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.simplemdm_push_certificate.test", "id", "push_certificate"),
					resource.TestCheckResourceAttrSet("data.simplemdm_push_certificate.test", "expires_at"),
				),
			},
		},
	})
}

func testAccPushCertificateDataSourceConfig() string {
	return `
data "simplemdm_push_certificate" "test" {}
`
}
