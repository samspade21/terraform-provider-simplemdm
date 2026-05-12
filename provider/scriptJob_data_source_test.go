package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccScriptJobDataSource self-provisions a script and a script job
// targeting the first enrolled device in the tenant, then reads it back via
// the data source. Skips cleanly when the tenant has no enrolled devices.
func TestAccScriptJobDataSource(t *testing.T) {
	testAccPreCheck(t)

	deviceID := findFirstDeviceID(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "simplemdm_script" "test" {
  name             = "tf_acc_scriptjob_data_source"
  variable_support = false
  content          = <<-EOT
		#!/bin/bash
		echo "tf-acc"
		EOT
}

resource "simplemdm_scriptjob" "test" {
  script_id  = simplemdm_script.test.id
  device_ids = ["%s"]
}

data "simplemdm_scriptjob" "test" {
  id = simplemdm_scriptjob.test.id
}
`, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.simplemdm_scriptjob.test", "id", "simplemdm_scriptjob.test", "id"),
					resource.TestCheckResourceAttrSet("data.simplemdm_scriptjob.test", "job_identifier"),
					resource.TestCheckResourceAttrSet("data.simplemdm_scriptjob.test", "status"),
					resource.TestCheckResourceAttrSet("data.simplemdm_scriptjob.test", "created_by"),
					resource.TestCheckResourceAttrSet("data.simplemdm_scriptjob.test", "variable_support"),
				),
			},
		},
	})
}
