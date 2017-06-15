package softlayer

import (
	"github.com/hashicorp/terraform/helper/resource"
	"regexp"
	"testing"
)

func TestAccSoftLayerSubnet_Basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerSubnetConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					// Check portable IP
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet", "type", "Portable"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet", "network", "PRIVATE"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet", "ip_version", "4"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet", "capacity", "4"),
					testAccCheckSoftLayerResources("softlayer_subnet.portable_subnet", "vlan_id",
						"softlayer_virtual_guest.subnetvm1", "private_vlan_id"),
					resource.TestMatchResourceAttr("softlayer_subnet.portable_subnet", "subnet",
						regexp.MustCompile(`^(([01]?[0-9]?[0-9]|2([0-4][0-9]|5[0-5]))\.){3}([01]?[0-9]?[0-9]|2([0-4][0-9]|5[0-5]))`)),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet", "notes", "portable_subnet"),
					// Check static IP
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet", "type", "Static"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet", "network", "PUBLIC"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet", "ip_version", "4"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet", "capacity", "4"),
					testAccCheckSoftLayerResources("softlayer_subnet.static_subnet", "endpoint_ip",
						"softlayer_virtual_guest.subnetvm1", "ipv4_address"),
					resource.TestMatchResourceAttr("softlayer_subnet.static_subnet", "subnet",
						regexp.MustCompile(`^(([01]?[0-9]?[0-9]|2([0-4][0-9]|5[0-5]))\.){3}([01]?[0-9]?[0-9]|2([0-4][0-9]|5[0-5]))`)),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet", "notes", "static_subnet"),
				),
			},

			resource.TestStep{
				Config: testAccCheckSoftLayerSubnetConfig_notes_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet", "notes", "portable_subnet_updated"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet", "notes", "static_subnet_updated"),
				),
			},
		},
	})
}

const testAccCheckSoftLayerSubnetConfig_basic = `
resource "softlayer_virtual_guest" "subnetvm1" {
    hostname = "subnetvm1"
    domain = "example.com"
    os_reference_code = "DEBIAN_7_64"
    datacenter = "dal06"
    network_speed = 100
    hourly_billing = true
    private_network_only = false
    cores = 1
    memory = 1024
    disks = [25]
    local_disk = false
}

resource "softlayer_subnet" "portable_subnet" {
  type = "Portable"
  network = "PRIVATE"
  ip_version = 4
  capacity = 4
  vlan_id = "${softlayer_virtual_guest.subnetvm1.private_vlan_id}"
  notes = "portable_subnet"
}

resource "softlayer_subnet" "static_subnet" {
  type = "Static"
  network = "PUBLIC"
  ip_version = 4
  capacity = 4
  endpoint_ip="${softlayer_virtual_guest.subnetvm1.ipv4_address}"
  notes = "static_subnet"
}
`

const testAccCheckSoftLayerSubnetConfig_notes_update = `
resource "softlayer_virtual_guest" "subnetvm1" {
    hostname = "subnetvm1"
    domain = "example.com"
    os_reference_code = "DEBIAN_7_64"
    datacenter = "dal06"
    network_speed = 100
    hourly_billing = true
    private_network_only = false
    cores = 1
    memory = 1024
    disks = [25]
    local_disk = false
}

resource "softlayer_subnet" "portable_subnet" {
  type = "Portable"
  network = "PRIVATE"
  ip_version = 4
  capacity = 4
  vlan_id = "${softlayer_virtual_guest.subnetvm1.private_vlan_id}"
  notes = "portable_subnet_updated"
}

resource "softlayer_subnet" "static_subnet" {
  type = "Static"
  network = "PUBLIC"
  ip_version = 4
  capacity = 4
  endpoint_ip="${softlayer_virtual_guest.subnetvm1.ipv4_address}"
  notes = "static_subnet_updated"
}`
