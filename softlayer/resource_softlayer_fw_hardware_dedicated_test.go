package softlayer

import (
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

func TestAccSoftLayerFwHardwareDedicated_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerFwHardwareDedicated_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated.accfw", "ha_enabled", "false"),
					testAccCheckSoftLayerResources("softlayer_fw_hardware_dedicated.accfw", "public_vlan_id",
						"softlayer_virtual_guest.fwvm1", "public_vlan_id"),
				),
			},
		},
	})
}

const testAccCheckSoftLayerFwHardwareDedicated_basic = `
resource "softlayer_virtual_guest" "fwvm1" {
    hostname = "fwvm1"
    domain = "example.com"
    os_reference_code = "DEBIAN_7_64"
    datacenter = "sjc01"
    network_speed = 10
    hourly_billing = true
    private_network_only = false
    cores = 1
    memory = 1024
    disks = [25]
    local_disk = false
}

resource "softlayer_fw_hardware_dedicated" "accfw" {
  ha_enabled = false
  public_vlan_id = "${softlayer_virtual_guest.fwvm1.public_vlan_id}"
}`
