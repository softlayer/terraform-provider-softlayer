package softlayer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
)

func TestAccSoftLayerLbVpxVip_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerLbVpxVipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSoftLayerLbVpxVipConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx_vip.testacc_vip", "load_balancing_method", "lc"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx_vip.testacc_vip", "name", "test_load_balancer_vip"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx_vip.testacc_vip", "source_port", "80"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx_vip.testacc_vip", "type", "HTTP"),
				),
			},
		},
	})
}

func testAccCheckSoftLayerLbVpxVipDestroy(s *terraform.State) error {
	service := services.GetNetworkApplicationDeliveryControllerService(testAccProvider.Meta().(*session.Session))

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "softlayer_lb_vpx_vip" {
			continue
		}

		nadcId, _ := strconv.Atoi(rs.Primary.Attributes["nad_controller_id"])
		vipName, _ := rs.Primary.Attributes["name"]

		vips, _ := service.
			Id(nadcId).
			Filter(filter.Path("name").Eq(vipName).Build()).
			GetLoadBalancers()

		if len(vips) > 0 {
			return fmt.Errorf("Netscaler VPX VIP still exists")
		}
	}

	return nil
}

var testAccCheckSoftLayerLbVpxVipConfig_basic = `


resource "softlayer_lb_vpx" "testacc_foobar_nadc" {
    datacenter = "dal09"
    speed = 10
    version = "10.1"
    plan = "Standard"
    ip_count = 2
}

resource "softlayer_lb_vpx_vip" "testacc_vip" {
    name = "test_load_balancer_vip"
    nad_controller_id = "${softlayer_lb_vpx.testacc_foobar_nadc.id}"
    load_balancing_method = "lc"
    source_port = 80
    type = "HTTP"
    virtual_ip_address = "${softlayer_lb_vpx.testacc_foobar_nadc.vip_pool[0]}"
}
`
