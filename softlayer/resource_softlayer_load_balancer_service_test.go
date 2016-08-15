package softlayer

import (
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

func TestAccSoftLayerLoadBalancerLocalService_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerLoadBalancerLocalServiceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"softlayer_lb_local_service.test_service", "port", "80"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_local_service.test_service", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_local_service.test_service", "weight", "1"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_local_service.test_service", "health_check_type", "DNS"),
				),
			},
		},
	})
}

const testAccCheckSoftLayerLoadBalancerLocalServiceConfig_basic = `
resource "softlayer_virtual_guest" "test_server_1" {
    name = "terraform-test"
    domain = "bar.example.com"
    image = "DEBIAN_7_64"
    region = "tok02"
    public_network_speed = 10
    hourly_billing = true
    private_network_only = false
    cpu = 1
    ram = 1024
    disks = [25, 10, 20]
    user_data = "{\"value\":\"newvalue\"}"
    dedicated_acct_host_only = true
    local_disk = false
}

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

resource "softlayer_lb_local_service" "test_service" {
    port = 80
    enabled = true
    service_group_id = "${softlayer_lb_local_service_group.test_service_group.service_group_id}"
    weight = 1
    health_check_type = "DNS"
    ip_address_id = "${softlayer_virtual_guest.test_server_1.ip_address_id}"
}
`
