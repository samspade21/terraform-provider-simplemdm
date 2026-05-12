package provider

import (
	"context"
	"fmt"
	"testing"

	simplemdm "github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckScriptJobDestroy(s *terraform.State) error {
	return testAccCheckResourceDestroyed("simplemdm_scriptjob", func(client *simplemdm.Client, id string) error {
		_, err := fetchScriptJobDetails(context.Background(), client, id)
		return err
	})(s)
}

// TestAccScriptJobResource targets a device group by ID. The script_jobs
// endpoint accepts assignment_group_ids only, and arbitrary auto-discovered
// legacy device groups don't qualify (422). Gate on SIMPLEMDM_DEVICE_GROUP_ID
// and skip otherwise.
func TestAccScriptJobResource(t *testing.T) {
	testAccPreCheck(t)

	deviceGroupID := testAccRequireEnv(t, "SIMPLEMDM_DEVICE_GROUP_ID")

	scriptResource := `
resource "simplemdm_script" "test_script" {
  name             = "tf_acc_scriptjob_resource_script"
  variable_support = true
  content          = <<-EOT
		#!/bin/bash
		echo "tf-acc"
		EOT
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckScriptJobDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + scriptResource + fmt.Sprintf(`
resource "simplemdm_scriptjob" "test_job" {
  script_id  = simplemdm_script.test_script.id
  device_ids = []
  group_ids  = [%q]
}
`, deviceGroupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "id"),
					resource.TestCheckResourceAttr("simplemdm_scriptjob.test_job", "group_ids.#", "1"),
					resource.TestCheckResourceAttr("simplemdm_scriptjob.test_job", "group_ids.0", deviceGroupID),
					resource.TestCheckResourceAttrPair(
						"simplemdm_scriptjob.test_job", "script_id",
						"simplemdm_script.test_script", "id",
					),
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "job_identifier"),
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "status"),
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "pending_count"),
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "created_at"),
					resource.TestCheckResourceAttrSet("simplemdm_scriptjob.test_job", "variable_support"),
				),
			},
		},
	})
}
