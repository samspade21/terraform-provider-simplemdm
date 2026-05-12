package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccPushCertificateResource verifies the upload path. Requires
// SIMPLEMDM_PUSH_CERT_PEM (path to a real signed APNs PEM file) and
// SIMPLEMDM_PUSH_CERT_APPLE_ID. Without these, the test is skipped because
// SimpleMDM cannot accept a synthetic certificate.
func TestAccPushCertificateResource(t *testing.T) {
	testAccPreCheck(t)
	pemPath := testAccRequireEnv(t, "SIMPLEMDM_PUSH_CERT_PEM")
	appleID := testAccGetEnv(t, "SIMPLEMDM_PUSH_CERT_APPLE_ID")

	if _, err := os.Stat(pemPath); err != nil {
		t.Skipf("Push certificate file at %s not accessible: %v", pemPath, err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "simplemdm_push_certificate" "apns" {
  certificate = file(%q)
  apple_id    = %q
}
`, pemPath, appleID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("simplemdm_push_certificate.apns", "expires_at"),
					resource.TestCheckResourceAttrSet("simplemdm_push_certificate.apns", "certificate_sha256"),
				),
			},
		},
	})
}
