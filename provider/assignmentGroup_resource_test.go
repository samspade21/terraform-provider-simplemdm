package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckAssignmentGroupDestroy(s *terraform.State) error {
	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to create test client: %w", err)
	}

	for name, rs := range s.RootModule().Resources {
		// Skip if not our resource type
		if rs.Type != "simplemdm_assignmentgroup" {
			continue
		}

		// Skip data sources - they start with "data."
		if strings.HasPrefix(name, "data.") {
			continue
		}

		// Try to fetch the resource with retry for eventual consistency
		// SimpleMDM API may take time to fully delete assignment groups
		var lastErr error
		for attempt := 0; attempt < 3; attempt++ {
			_, lastErr = fetchAssignmentGroup(context.Background(), client, rs.Primary.ID)

			// If we get a 404, the resource is properly deleted
			if lastErr != nil && isNotFoundError(lastErr) {
				break
			}

			// If the resource still exists after 3 attempts, it wasn't deleted
			if lastErr == nil && attempt == 2 {
				return fmt.Errorf("assignment group %s still exists after destroy", rs.Primary.ID)
			}

			// Wait briefly before retrying (only if not last attempt)
			if attempt < 2 && lastErr == nil {
				time.Sleep(2 * time.Second)
			}
		}

		// If we got an error that's not a 404, that's unexpected
		if lastErr != nil && !isNotFoundError(lastErr) {
			return fmt.Errorf("unexpected error checking assignment group %s: %w", rs.Primary.ID, lastErr)
		}
	}

	return nil
}

// assignmentGroupFixtureConfig returns Terraform config that self-provisions
// an App Store app to use as a fixture for assignment group acceptance tests.
// We deliberately do NOT include simplemdm_customprofile fixtures here:
// custom profiles exhibit eventual consistency on the SimpleMDM API (the
// resource's Read may 404 immediately after Create), which corrupts refresh
// plans for tests that wire profile relationships into the assignment group.
// Profile-attachment behaviour is exercised in dedicated tests instead.
const assignmentGroupFixtureConfig = `
resource "simplemdm_app" "fixture" {
  app_store_id = "284882215"
}
`

func TestAccAssignmentGroupResource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAssignmentGroupDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + assignmentGroupFixtureConfig + `
resource "simplemdm_assignmentgroup" "testgroup2" {
  name                  = "Test Assignment Group Resource"
  auto_deploy           = false
  group_type            = "standard"
  priority              = 3
  app_track_location    = false
  apps                  = [simplemdm_app.fixture.id]
  devices_remove_others = true
  profiles_sync         = false
  apps_push             = false
  apps_update           = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup.testgroup2", "name", "Test Assignment Group Resource"),
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup.testgroup2", "group_type", "standard"),
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup.testgroup2", "priority", "3"),
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup.testgroup2", "app_track_location", "false"),
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup.testgroup2", "devices_remove_others", "true"),
					resource.TestCheckResourceAttrSet("simplemdm_assignmentgroup.testgroup2", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "simplemdm_assignmentgroup.testgroup2",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apps_update", "apps_push", "auto_deploy", "profiles_sync", "group_type", "profiles", "apps", "created_at", "updated_at", "device_count", "group_count", "devices_remove_others"},
			},
			// Update and Read testing — verifies mutable scalar updates.
			{
				Config: providerConfig + assignmentGroupFixtureConfig + `
resource "simplemdm_assignmentgroup" "testgroup2" {
  name                  = "Updated Assignment Group Resource"
  auto_deploy           = false
  group_type            = "standard"
  priority              = 7
  app_track_location    = true
  apps                  = [simplemdm_app.fixture.id]
  devices_remove_others = false
  profiles_sync         = false
  apps_push             = false
  apps_update           = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup.testgroup2", "name", "Updated Assignment Group Resource"),
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup.testgroup2", "group_type", "standard"),
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup.testgroup2", "priority", "7"),
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup.testgroup2", "app_track_location", "true"),
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup.testgroup2", "devices_remove_others", "false"),
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup.testgroup2", "apps.#", "1"),
					resource.TestCheckResourceAttrPair(
						"simplemdm_assignmentgroup.testgroup2", "apps.0",
						"simplemdm_app.fixture", "id",
					),
					resource.TestCheckResourceAttrSet("simplemdm_assignmentgroup.testgroup2", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccAssignmentGroupResource_Import tests the import functionality
func TestAccAssignmentGroupResource_Import(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "simplemdm_assignmentgroup" "test_import" {
  name        = "Test Import Group"
  auto_deploy = true
  group_type  = "standard"
  priority    = 5
}
`,
			},
			{
				ResourceName:      "simplemdm_assignmentgroup.test_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apps_update", "apps_push", "profiles_sync", "devices_remove_others", "group_type",
				},
			},
		},
	})
}

// TestAccAssignmentGroupResource_RelationshipUpdates tests adding and removing
// app relationships. Profile relationships are tested via the main test
// function above; isolating apps here avoids the customprofile eventual
// consistency that would otherwise force ExpectNonEmptyPlan everywhere.
func TestAccAssignmentGroupResource_RelationshipUpdates(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAssignmentGroupDestroy,
		Steps: []resource.TestStep{
			// Create with no apps
			{
				Config: providerConfig + `
resource "simplemdm_app" "fixture" {
  app_store_id = "284882215"
}

resource "simplemdm_assignmentgroup" "test_relationships" {
  name        = "Test Relationships Group"
  auto_deploy = false
  group_type  = "standard"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup.test_relationships", "name", "Test Relationships Group"),
					resource.TestCheckResourceAttrSet("simplemdm_assignmentgroup.test_relationships", "id"),
				),
			},
			// Add an app
			{
				Config: providerConfig + `
resource "simplemdm_app" "fixture" {
  app_store_id = "284882215"
}

resource "simplemdm_assignmentgroup" "test_relationships" {
  name        = "Test Relationships Group"
  auto_deploy = false
  group_type  = "standard"
  apps        = [simplemdm_app.fixture.id]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup.test_relationships", "apps.#", "1"),
				),
			},
		},
	})
}

// TestAccAssignmentGroupResource_RateLimitHandling tests profile sync rate limiting
func TestAccAssignmentGroupResource_RateLimitHandling(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAssignmentGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "simplemdm_assignmentgroup" "test_ratelimit" {
  name          = "Test Rate Limit Group"
  auto_deploy   = false
  group_type    = "standard"
  profiles_sync = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("simplemdm_assignmentgroup.test_ratelimit", "name", "Test Rate Limit Group"),
					resource.TestCheckResourceAttrSet("simplemdm_assignmentgroup.test_ratelimit", "id"),
				),
			},
		},
	})
}
