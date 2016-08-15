package softlayer

import (
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

func TestAccSoftLayerLoadBalancerLocalServiceGroup_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerLoadBalancerServiceGroupConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"softlayer_lb_local_service_group.test_service_group", "port", "82"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_local_service_group.test_service_group", "routing_method", "CONSISTENT_HASH_IP"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_local_service_group.test_service_group", "routing_type", "HTTP"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_local_service_group.test_service_group", "allocation", "100"),
				),
			},
		},
	})
}

const testAccCheckSoftLayerLoadBalancerServiceGroupConfig_basic = `
resource "softlayer_lb_local" "testacc_foobar_lb" {
    connections = 15000
    location    = "tok02"
    ha_enabled  = false
}

resource "softlayer_lb_local_service_group" "test_service_group" {
    port = 82
    routing_method = "CONSISTENT_HASH_IP"
    routing_type = "HTTP"
    load_balancer_id = "${softlayer_lb_local.testacc_foobar_lb.id}"
    allocation = 100
}
`
