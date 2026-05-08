package provider

import (
	"context"
	"testing"

	simplemdm "github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckEnrollmentDestroy(s *terraform.State) error {
	return testAccCheckResourceDestroyed("simplemdm_enrollment", func(client *simplemdm.Client, id string) error {
		_, err := fetchEnrollment(context.Background(), client, id)
		return err
	})(s)
}

// TestAccEnrollmentResource creates an enrollment against an existing legacy
// device group, exercising the invitation update path with hardcoded test
// contacts. The /enrollments endpoint rejects both freshly-created assignment
// groups and arbitrary legacy device groups with 404, so the device group ID
// must be supplied explicitly via SIMPLEMDM_DEVICE_GROUP_ID. The test skips
// cleanly when that var isn't set.
func TestAccEnrollmentResource(t *testing.T) {
	testAccPreCheck(t)

	deviceGroupID := testAccRequireEnv(t, "SIMPLEMDM_DEVICE_GROUP_ID")

	const (
		initialContact = "tf-acc-test@example.com"
		updatedContact = "tf-acc-test-updated@example.com"
	)

	steps := []resource.TestStep{
		{
			Config: providerConfig + `
resource "simplemdm_enrollment" "test" {
  device_group_id    = "` + deviceGroupID + `"
  user_enrollment    = false
  welcome_screen     = true
  authentication     = false
  invitation_contact = "` + initialContact + `"
}
`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("simplemdm_enrollment.test", "user_enrollment", "false"),
				resource.TestCheckResourceAttr("simplemdm_enrollment.test", "device_group_id", deviceGroupID),
				resource.TestCheckResourceAttr("simplemdm_enrollment.test", "welcome_screen", "true"),
				resource.TestCheckResourceAttr("simplemdm_enrollment.test", "authentication", "false"),
				resource.TestCheckResourceAttr("simplemdm_enrollment.test", "invitation_contact", initialContact),
				resource.TestCheckResourceAttrSet("simplemdm_enrollment.test", "id"),
			),
		},
		{
			ResourceName:      "simplemdm_enrollment.test",
			ImportState:       true,
			ImportStateVerify: true,
			ImportStateVerifyIgnore: []string{
				"invitation_contact",
			},
		},
		{
			Config: providerConfig + `
resource "simplemdm_enrollment" "test" {
  device_group_id    = "` + deviceGroupID + `"
  user_enrollment    = false
  welcome_screen     = true
  authentication     = false
  invitation_contact = "` + updatedContact + `"
}
`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("simplemdm_enrollment.test", "device_group_id", deviceGroupID),
				resource.TestCheckResourceAttr("simplemdm_enrollment.test", "invitation_contact", updatedContact),
			),
		},
		{
			Config: providerConfig + `
resource "simplemdm_enrollment" "test" {
  device_group_id = "` + deviceGroupID + `"
  user_enrollment = false
  welcome_screen  = true
  authentication  = false
}
`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckNoResourceAttr("simplemdm_enrollment.test", "invitation_contact"),
			),
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckEnrollmentDestroy,
		Steps:                    steps,
	})
}
