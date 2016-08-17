package softlayer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.ibm.com/riethm/gopherlayer.git/datatypes"
	"github.ibm.com/riethm/gopherlayer.git/services"
	"github.ibm.com/riethm/gopherlayer.git/session"
)

func TestAccSoftLayerProvisioningHook_Basic(t *testing.T) {
	var hook datatypes.Provisioning_Hook

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerProvisioningHookDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerProvisioningHookConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerProvisioningHookExists("softlayer_provisioning_hook.test-provisioning-hook", &hook),
					testAccCheckSoftLayerProvisioningHookAttributes(&hook),
					resource.TestCheckResourceAttr(
						"softlayer_provisioning_hook.test-provisioning-hook", "name", "test-sl-hook"),
					resource.TestCheckResourceAttr(
						"softlayer_provisioning_hook.test-provisioning-hook", "uri", "http://www.weather.com"),
				),
			},

			resource.TestStep{
				Config: testAccCheckSoftLayerProvisioningHookConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerProvisioningHookExists("softlayer_provisioning_hook.updated-test-provisioning-hook", &hook),
					resource.TestCheckResourceAttr(
						"softlayer_provisioning_hook.updated-test-provisioning-hook", "name", "changed_name"),
					resource.TestCheckResourceAttr(
						"softlayer_provisioning_hook.updated-test-provisioning-hook", "uri", "https://www.github.com"),
				),
			},
		},
	})
}

func testAccCheckSoftLayerProvisioningHookDestroy(s *terraform.State) error {
	service := services.GetProvisioningHookService(testAccProvider.Meta().(*session.Session))

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "softlayer_provisioning_hook" {
			continue
		}

		hookId, _ := strconv.Atoi(rs.Primary.ID)

		// Try to find the provisioning hook
		_, err := service.Id(hookId).GetObject()

		if err == nil {
			return fmt.Errorf("Provisioning Hook still exists")
		}
	}

	return nil
}

func testAccCheckSoftLayerProvisioningHookAttributes(hook *datatypes.Provisioning_Hook) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if hook.Name != "test-sl-hook" {
			return fmt.Errorf("Bad name: %s", hook.Name)
		}

		if hook.Uri != "http://www.weather.com" {
			return fmt.Errorf("Bad uri: %s", hook.Uri)
		}

		return nil
	}
}

func testAccCheckSoftLayerProvisioningHookExists(n string, hook *datatypes.Provisioning_Hook) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		hookId, _ := strconv.Atoi(rs.Primary.ID)

		service := services.GetProvisioningHookService(testAccProvider.Meta().(*session.Session))
		foundHook, err := service.Id(hookId).GetObject()

		if err != nil {
			return err
		}

		if strconv.Itoa(int(foundHook.Id)) != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		*hook = foundHook

		return nil
	}
}

const testAccCheckSoftLayerProvisioningHookConfig_basic = `
resource "softlayer_provisioning_hook" "test-provisioning-hook" {
    name = "test-sl-hook"
    uri = "http://www.weather.com"
}`

const testAccCheckSoftLayerProvisioningHookConfig_updated = `
resource "softlayer_provisioning_hook" "updated-test-provisioning-hook" {
    name = "changed_name"
    uri  = "https://www.github.com"
}`
