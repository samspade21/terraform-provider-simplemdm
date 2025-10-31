package provider

import (
	"context"
	"os"
	"testing"

	simplemdm "github.com/DavidKrau/simplemdm-go-client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckEnrollmentDestroy(s *terraform.State) error {
	return testAccCheckResourceDestroyed("simplemdm_enrollment", func(client *simplemdm.Client, id string) error {
		_, err := fetchEnrollment(context.Background(), client, id)
		return err
	})(s)
}

func TestAccEnrollmentResource(t *testing.T) {
	testAccPreCheck(t)

	// Get the required contact email (still needed as it's user-specific)
	invitationContact := testAccRequireEnv(t, "SIMPLEMDM_ENROLLMENT_CONTACT")

	steps := []resource.TestStep{
		{
			Config: providerConfig + `
				# Create prerequisite assignment group (device groups cannot be created via API)
				resource "simplemdm_assignmentgroup" "test_group" {
					name                  = "Test Enrollment Assignment Group"
					auto_deploy_profiles  = false
					auto_deploy_apps      = false
				}

				# Create enrollment using dynamic reference
				resource "simplemdm_enrollment" "test" {
					device_group_id    = simplemdm_assignmentgroup.test_group.id
					user_enrollment    = false
					welcome_screen     = true
					authentication     = false
					invitation_contact = "` + invitationContact + `"

					depends_on = [simplemdm_assignmentgroup.test_group]
				}
			`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("simplemdm_enrollment.test", "user_enrollment", "false"),
				resource.TestCheckResourceAttr("simplemdm_enrollment.test", "welcome_screen", "true"),
				resource.TestCheckResourceAttr("simplemdm_enrollment.test", "authentication", "false"),
				resource.TestCheckResourceAttr("simplemdm_enrollment.test", "invitation_contact", invitationContact),
				// Verify dynamic relationship
				resource.TestCheckResourceAttrPair(
					"simplemdm_enrollment.test", "device_group_id",
					"simplemdm_assignmentgroup.test_group", "id",
				),
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
	}

	// Optional: test updating the invitation contact if provided
	if updatedContact := os.Getenv("SIMPLEMDM_ENROLLMENT_CONTACT_UPDATE"); updatedContact != "" {
		steps = append(steps, resource.TestStep{
			Config: providerConfig + `
				# Keep the same assignment group
				resource "simplemdm_assignmentgroup" "test_group" {
					name                  = "Test Enrollment Assignment Group"
					auto_deploy_profiles  = false
					auto_deploy_apps      = false
				}

				# Update enrollment with new contact
				resource "simplemdm_enrollment" "test" {
					device_group_id    = simplemdm_assignmentgroup.test_group.id
					user_enrollment    = false
					welcome_screen     = true
					authentication     = false
					invitation_contact = "` + updatedContact + `"

					depends_on = [simplemdm_assignmentgroup.test_group]
				}
			`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrPair(
					"simplemdm_enrollment.test", "device_group_id",
					"simplemdm_assignmentgroup.test_group", "id",
				),
				resource.TestCheckResourceAttr("simplemdm_enrollment.test", "invitation_contact", updatedContact),
			),
		})
	}

	// Test removing the invitation contact
	steps = append(steps, resource.TestStep{
		Config: providerConfig + `
			# Keep the same assignment group
			resource "simplemdm_assignmentgroup" "test_group" {
				name                  = "Test Enrollment Assignment Group"
				auto_deploy_profiles  = false
				auto_deploy_apps      = false
			}

			# Remove invitation contact
			resource "simplemdm_enrollment" "test" {
				device_group_id = simplemdm_assignmentgroup.test_group.id
				user_enrollment = false
				welcome_screen  = true
				authentication  = false

				depends_on = [simplemdm_assignmentgroup.test_group]
			}
		`,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckNoResourceAttr("simplemdm_enrollment.test", "invitation_contact"),
		),
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckEnrollmentDestroy,
		Steps:                    steps,
	})
}
