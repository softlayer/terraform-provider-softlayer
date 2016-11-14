package softlayer

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccSoftLayerLbVpxHa_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSoftLayerLbVpxHaConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx_ha.test_ha", "stay_secondary", "true"),
					testAccCheckSoftLayerResources("softlayer_lb_vpx_ha.test_ha", "primary_id",
						"softlayer_lb_vpx.test_pri", "id"),
					testAccCheckSoftLayerResources("softlayer_lb_vpx_ha.test_ha", "secondary_id",
						"softlayer_lb_vpx.test_sec", "id"),
				),
			},
		},
	})
}

var testAccCheckSoftLayerLbVpxHaConfig_basic = `

resource "softlayer_virtual_guest" "vm1" {
    hostname = "vm1"
    domain = "example.com"
    os_reference_code = "DEBIAN_7_64"
    datacenter = "tok02"
    network_speed = 10
    hourly_billing = true
    private_network_only = false
    cores = 1
    memory = 1024
    disks = [25]
    local_disk = false
}

resource "softlayer_lb_vpx" "test_pri" {
    datacenter = "tok02"
    speed = 10
    version = "10.5"
    plan = "Standard"
    ip_count = 2
        public_vlan_id = "${softlayer_virtual_guest.vm1.public_vlan_id}"
    private_vlan_id = "${softlayer_virtual_guest.vm1.private_vlan_id}"
    public_subnet = "${softlayer_virtual_guest.vm1.public_subnet}"
    private_subnet = "${softlayer_virtual_guest.vm1.private_subnet}"
}

resource "softlayer_lb_vpx" "test_sec" {
    datacenter = "tok02"
    speed = 10
    version = "10.5"
    plan = "Standard"
    ip_count = 2
    public_vlan_id = "${softlayer_virtual_guest.vm1.public_vlan_id}"
    private_vlan_id = "${softlayer_virtual_guest.vm1.private_vlan_id}"
    public_subnet = "${softlayer_virtual_guest.vm1.public_subnet}"
    private_subnet = "${softlayer_virtual_guest.vm1.private_subnet}"
}

resource "softlayer_lb_vpx_ha" "test_ha" {
    primary_id = "${softlayer_lb_vpx.test_pri.id}"
    secondary_id = "${softlayer_lb_vpx.test_sec.id}"
    stay_secondary = true
}
`
