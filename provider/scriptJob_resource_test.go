package provider

import (
	"context"
	"testing"

	simplemdm "github.com/DavidKrau/simplemdm-go-client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckScriptJobDestroy(s *terraform.State) error {
	return testAccCheckResourceDestroyed("simplemdm_scriptjob", func(client *simplemdm.Client, id string) error {
		_, err := fetchScriptJobDetails(context.Background(), client, id)
		return err
	})(s)
}

func TestAccScriptJobResource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckScriptJobDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
					# Create dynamic script for testing
					resource "simplemdm_script" "test_script" {
						name            = "Test Script Job Script"
						scriptfile      = file("./testfiles/testscript.sh")
						variablesupport = true
					}

					# Create dynamic device group for testing
					resource "simplemdm_devicegroup" "test_group" {
						name = "Test Script Job Device Group"
					}

					# Create script job using dynamic resources
					resource "simplemdm_scriptjob" "test_job" {
						script_id  = simplemdm_script.test_script.id
						device_ids = []
						group_ids  = [simplemdm_devicegroup.test_group.id]
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check the script job attributes
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "id"),
					resource.TestCheckResourceAttr("simplemdm_scriptjob.test_job", "group_ids.#", "1"),
					// Verify dynamic relationships
					resource.TestCheckResourceAttrPair(
						"simplemdm_scriptjob.test_job", "script_id",
						"simplemdm_script.test_script", "id",
					),
					resource.TestCheckResourceAttrPair(
						"simplemdm_scriptjob.test_job", "group_ids.0",
						"simplemdm_devicegroup.test_group", "id",
					),
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "job_identifier"),
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "status"),
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "pending_count"),
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "created_at"),
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "variable_support"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "simplemdm_scriptjob.test_job",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
					# Keep the same script
					resource "simplemdm_script" "test_script" {
						name            = "Test Script Job Script"
						scriptfile      = file("./testfiles/testscript.sh")
						variablesupport = true
					}

					# Keep the same device group
					resource "simplemdm_devicegroup" "test_group" {
						name = "Test Script Job Device Group"
					}

					# Update script job with custom attributes
					resource "simplemdm_scriptjob" "test_job" {
						script_id              = simplemdm_script.test_script.id
						device_ids             = []
						group_ids              = [simplemdm_devicegroup.test_group.id]
						custom_attribute       = "SomeAttribute"
						custom_attribute_regex = ".*"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check the updated script job attributes
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "id"),
					resource.TestCheckResourceAttr("simplemdm_scriptjob.test_job", "custom_attribute", "SomeAttribute"),
					resource.TestCheckResourceAttr("simplemdm_scriptjob.test_job", "custom_attribute_regex", ".*"),
					resource.TestCheckResourceAttr("simplemdm_scriptjob.test_job", "group_ids.#", "1"),
					// Verify dynamic relationships
					resource.TestCheckResourceAttrPair(
						"simplemdm_scriptjob.test_job", "script_id",
						"simplemdm_script.test_script", "id",
					),
					resource.TestCheckResourceAttrPair(
						"simplemdm_scriptjob.test_job", "group_ids.0",
						"simplemdm_devicegroup.test_group", "id",
					),
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "status"),
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "job_identifier"),
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "success_count"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
