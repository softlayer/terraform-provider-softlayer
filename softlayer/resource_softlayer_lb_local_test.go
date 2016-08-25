package softlayer

import (
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

func TestAccSoftLayerLbLocal_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerLbLocalConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"softlayer_lb_local.testacc_foobar_lb", "connections", "15000"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_local.testacc_foobar_lb", "location", "tok02"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_local.testacc_foobar_lb", "ha_enabled", "false"),
				),
			},
		},
	})
}

const testAccCheckSoftLayerLbLocalConfig_basic = `
resource "softlayer_lb_local" "testacc_foobar_lb" {
    connections = 15000
    location    = "tok02"
    ha_enabled  = false
}`
