package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScriptDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "simplemdm_script" "test" {
  name             = "tf_acc_test_script"
  variable_support = false
  content          = <<-EOT
		#!/bin/bash
		echo "hello from acceptance test"
		EOT
}

data "simplemdm_script" "test" {
  id = simplemdm_script.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.simplemdm_script.test", "id", "simplemdm_script.test", "id"),
					resource.TestCheckResourceAttr("data.simplemdm_script.test", "name", "tf_acc_test_script"),
					resource.TestCheckResourceAttrSet("data.simplemdm_script.test", "content"),
					resource.TestCheckResourceAttrSet("data.simplemdm_script.test", "variable_support"),
					resource.TestCheckResourceAttrSet("data.simplemdm_script.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.simplemdm_script.test", "updated_at"),
				),
			},
		},
	})
}
