package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckAttributeDestroy(s *terraform.State) error {
	client, err := getTestClient()
	if err != nil {
		return err
	}

	// Check only the resources that remain in the final state
	// When an attribute name changes, Terraform replaces the resource (delete old, create new)
	// The old attribute is deleted during the replacement, so we only check what's in final state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "simplemdm_attribute" {
			continue
		}

		// The attribute ID is the attribute name
		attributeName := rs.Primary.ID

		// Try to fetch the resource with retry for eventual consistency
		// SimpleMDM API may take time to fully delete attributes
		var lastErr error
		for attempt := 0; attempt < 3; attempt++ {
			_, lastErr = client.AttributeGet(attributeName)

			// If we get a 404, the resource is properly deleted
			if lastErr != nil && isNotFoundError(lastErr) {
				break
			}

			// If the resource still exists after 3 attempts, it wasn't deleted
			if lastErr == nil && attempt == 2 {
				return fmt.Errorf("attribute %s still exists after destroy", attributeName)
			}

			// Wait briefly before retrying (only if not last attempt)
			if attempt < 2 && lastErr == nil {
				time.Sleep(2 * time.Second)
			}
		}

		// If we got an error that's not a 404, that's unexpected
		if lastErr != nil && !isNotFoundError(lastErr) {
			return fmt.Errorf("unexpected error checking attribute %s: %w", attributeName, lastErr)
		}
	}

	return nil
}

func TestAccAttributeResource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAttributeDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
				resource "simplemdm_attribute" "testattribute" {
					name= "newAttribute"
					default_value= "test value for test attribute"
				  }
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify attributes
					resource.TestCheckResourceAttr("simplemdm_attribute.testattribute", "name", "newAttribute"),
					resource.TestCheckResourceAttr("simplemdm_attribute.testattribute", "default_value", "test value for test attribute"),
					resource.TestCheckResourceAttr("simplemdm_attribute.testattribute", "id", "newAttribute"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "simplemdm_attribute.testattribute",
				ImportState:       true,
				ImportStateVerify: true,
				//ImportStateVerifyIgnore: []string{"filesha", "mobileconfig"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
				resource "simplemdm_attribute" "testattribute" {
					name= "newAttribute2"
					default_value= ""
				  }
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify attributes
					resource.TestCheckResourceAttr("simplemdm_attribute.testattribute", "name", "newAttribute2"),
					resource.TestCheckResourceAttr("simplemdm_attribute.testattribute", "default_value", ""),
					resource.TestCheckResourceAttr("simplemdm_attribute.testattribute", "id", "newAttribute2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
