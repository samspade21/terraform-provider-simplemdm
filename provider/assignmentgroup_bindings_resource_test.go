package provider

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccDynamicAssignmentGroupID returns the assignment group ID to bind
// against. Honours SIMPLEMDM_TEST_DYNAMIC_GROUP_ID for explicit pinning;
// otherwise discovers the first dynamic assignment group in the tenant.
// Skips the test if neither is available — the binding tests can only run
// against an existing dynamic AG (the binding resources don't own the AG).
func testAccDynamicAssignmentGroupID(t *testing.T) string {
	t.Helper()
	return envOrDiscover(t, "SIMPLEMDM_TEST_DYNAMIC_GROUP_ID", "a dynamic assignment group", func() (string, error) {
		client, err := getTestClient()
		if err != nil {
			return "", err
		}
		groups, err := fetchAllAssignmentGroups(context.Background(), client)
		if err != nil {
			return "", err
		}
		for _, g := range groups {
			if g.Attributes.GroupType == "dynamic" {
				return fmt.Sprintf("%d", g.ID), nil
			}
		}
		return "", nil
	})
}

// testAccCheckAssignmentGroupStillExists confirms that, after destroying the
// binding resources, the parent assignment group is still around (i.e. our
// binding resources didn't accidentally take ownership). Also verifies the
// group's group_type to confirm dynamic groups remain dynamic.
func testAccCheckAssignmentGroupStillExists(groupID, wantGroupType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client, err := getTestClient()
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}
		ag, err := fetchAssignmentGroup(context.Background(), client, groupID)
		if err != nil {
			return fmt.Errorf("expected assignment group %s to still exist after destroy: %w", groupID, err)
		}
		if got := ag.Data.Attributes.GroupType; got != wantGroupType {
			return fmt.Errorf("assignment group %s: expected group_type %q, got %q", groupID, wantGroupType, got)
		}
		return nil
	}
}

// testAccCheckProfileBindingAbsent verifies the destroy path actually removed
// the join. Uses the same inverse lookup the resource's Read uses.
func testAccCheckProfileBindingAbsent(s *terraform.State) error {
	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to create test client: %w", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "simplemdm_assignmentgroup_profile_binding" {
			continue
		}
		profileID := rs.Primary.Attributes["profile_id"]
		agID := rs.Primary.Attributes["assignment_group_id"]

		url := fmt.Sprintf("https://%s/api/v1/profiles/%s", client.HostName, profileID)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		body, err := client.RequestResponse200(req)
		if err != nil {
			if isNotFoundError(err) {
				// Profile got deleted (probably by the customprofile resource
				// in the same plan). That implicitly removes the binding.
				continue
			}
			return fmt.Errorf("unexpected error checking profile %s: %w", profileID, err)
		}
		bound, err := profileHasGroupAssignment(body, agID, profileID)
		if err != nil {
			return fmt.Errorf("error parsing profile %s: %w", profileID, err)
		}
		if bound {
			return fmt.Errorf("profile binding %s -> %s still present after destroy", profileID, agID)
		}
	}
	return nil
}

// testAccCheckAppBindingAbsent uses the inverse lookup on the assignment group
// (apps don't expose relationships) to confirm the join is gone.
func testAccCheckAppBindingAbsent(s *terraform.State) error {
	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to create test client: %w", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "simplemdm_assignmentgroup_app_binding" {
			continue
		}
		appID := rs.Primary.Attributes["app_id"]
		agID := rs.Primary.Attributes["assignment_group_id"]

		url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s", client.HostName, agID)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		body, err := client.RequestResponse200(req)
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return fmt.Errorf("unexpected error checking assignment group %s: %w", agID, err)
		}
		bound, err := assignmentGroupHasAppBinding(body, appID, agID)
		if err != nil {
			return fmt.Errorf("error parsing assignment group %s: %w", agID, err)
		}
		if bound {
			return fmt.Errorf("app binding %s -> %s still present after destroy", appID, agID)
		}
	}
	return nil
}

// TestAccAssignmentGroupProfileBinding creates a fresh custom configuration
// profile, binds it to the dynamic assignment group, and confirms the binding
// is visible via the inverse lookup. On destroy it verifies the binding is
// gone but the assignment group itself is untouched.
func TestAccAssignmentGroupProfileBinding(t *testing.T) {
	testAccPreCheck(t)

	agID := testAccDynamicAssignmentGroupID(t)

	config := providerConfig + fmt.Sprintf(`
resource "simplemdm_customprofile" "binding_fixture" {
  name                      = "tf_acc_assignmentgroup_profile_binding_fixture"
  mobileconfig              = file("./testfiles/firewall-test-profile.mobileconfig")
  user_scope                = false
  attribute_support         = false
  escape_attributes         = false
  reinstall_after_os_update = false
}

resource "simplemdm_assignmentgroup_profile_binding" "test" {
  assignment_group_id = "%s"
  profile_id          = simplemdm_customprofile.binding_fixture.id
}
`, agID)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckProfileBindingAbsent,
			testAccCheckCustomProfileDestroy,
			// The parent AG must still exist (we don't own it). Validate the
			// group_type round-trips so we're sure we didn't accidentally
			// mutate the group's attributes.
			testAccCheckAssignmentGroupStillExists(agID, "dynamic"),
		),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup_profile_binding.test", "assignment_group_id", agID),
					resource.TestCheckResourceAttrPair(
						"simplemdm_assignmentgroup_profile_binding.test", "profile_id",
						"simplemdm_customprofile.binding_fixture", "id",
					),
					resource.TestCheckResourceAttrSet("simplemdm_assignmentgroup_profile_binding.test", "id"),
				),
			},
			{
				ResourceName:      "simplemdm_assignmentgroup_profile_binding.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccAssignmentGroupAppBinding self-provisions an App Store app (Apple's
// own "Pages", store id 361309726, which is freely available and not used by
// other acceptance tests in this package — avoids the cross-test fixture
// collision the simplemdm_app deletion eventual-consistency window can
// otherwise cause). The binding is created with default
// deployment_type/install_type, then updated to "standard" deployment to
// exercise the Update path.
func TestAccAssignmentGroupAppBinding(t *testing.T) {
	testAccPreCheck(t)

	agID := testAccDynamicAssignmentGroupID(t)

	baseFixture := `
resource "simplemdm_app" "binding_fixture" {
  app_store_id = "361309726"
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckAppBindingAbsent,
			testAccCheckAssignmentGroupStillExists(agID, "dynamic"),
		),
		Steps: []resource.TestStep{
			// Step 1: bind without overrides.
			{
				Config: providerConfig + baseFixture + fmt.Sprintf(`
resource "simplemdm_assignmentgroup_app_binding" "test" {
  assignment_group_id = "%s"
  app_id              = simplemdm_app.binding_fixture.id
}
`, agID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup_app_binding.test", "assignment_group_id", agID),
					resource.TestCheckResourceAttrPair(
						"simplemdm_assignmentgroup_app_binding.test", "app_id",
						"simplemdm_app.binding_fixture", "id",
					),
					resource.TestCheckResourceAttrSet("simplemdm_assignmentgroup_app_binding.test", "id"),
				),
			},
			// Step 2: import. ImportStateVerifyIgnore deployment_type /
			// install_type because the API does not return per-app override
			// values, so the imported state can never reproduce them.
			{
				ResourceName:            "simplemdm_assignmentgroup_app_binding.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deployment_type", "install_type"},
			},
			// Step 3: update deployment_type to "standard". This triggers the
			// unassign+reassign cycle in Update.
			{
				Config: providerConfig + baseFixture + fmt.Sprintf(`
resource "simplemdm_assignmentgroup_app_binding" "test" {
  assignment_group_id = "%s"
  app_id              = simplemdm_app.binding_fixture.id
  deployment_type     = "standard"
}
`, agID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup_app_binding.test", "deployment_type", "standard"),
					resource.TestCheckResourceAttrPair(
						"simplemdm_assignmentgroup_app_binding.test", "app_id",
						"simplemdm_app.binding_fixture", "id",
					),
				),
			},
		},
	})
}
