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
					// Check portable IPv4
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet", "type", "Portable"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet", "private", "true"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet", "ip_version", "4"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet", "capacity", "4"),
					testAccCheckSoftLayerResources("softlayer_subnet.portable_subnet", "vlan_id",
						"softlayer_virtual_guest.subnetvm1", "private_vlan_id"),
					resource.TestMatchResourceAttr("softlayer_subnet.portable_subnet", "subnet_cidr",
						regexp.MustCompile(`^(([01]?[0-9]?[0-9]|2([0-4][0-9]|5[0-5]))\.){3}([01]?[0-9]?[0-9]|2([0-4][0-9]|5[0-5]))`)),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet", "notes", "portable_subnet"),

					// Check static IPv4
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet", "type", "Static"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet", "private", "false"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet", "ip_version", "4"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet", "capacity", "4"),
					testAccCheckSoftLayerResources("softlayer_subnet.static_subnet", "endpoint_ip",
						"softlayer_virtual_guest.subnetvm1", "ipv4_address"),
					resource.TestMatchResourceAttr("softlayer_subnet.static_subnet", "subnet_cidr",
						regexp.MustCompile(`^(([01]?[0-9]?[0-9]|2([0-4][0-9]|5[0-5]))\.){3}([01]?[0-9]?[0-9]|2([0-4][0-9]|5[0-5]))`)),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet", "notes", "static_subnet"),

					// Check portable IPv6
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet_v6", "type", "Portable"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet_v6", "private", "false"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet_v6", "ip_version", "6"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet_v6", "capacity", "64"),
					testAccCheckSoftLayerResources("softlayer_subnet.portable_subnet_v6", "vlan_id",
						"softlayer_virtual_guest.subnetvm1", "public_vlan_id"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.portable_subnet_v6", "notes", "portable_subnet"),
					// Check static IPv6
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet_v6", "type", "Static"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet_v6", "private", "false"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet_v6", "ip_version", "6"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet_v6", "capacity", "64"),
					testAccCheckSoftLayerResources("softlayer_subnet.static_subnet_v6", "endpoint_ip",
						"softlayer_virtual_guest.subnetvm1", "ipv6_address"),
					resource.TestCheckResourceAttr(
						"softlayer_subnet.static_subnet_v6", "notes", "static_subnet"),
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
    datacenter = "wdc04"
    network_speed = 100
    hourly_billing = true
    private_network_only = false
    cores = 1
    memory = 1024
    disks = [25]
    local_disk = false
    ipv6_enabled = true
}

resource "softlayer_subnet" "portable_subnet" {
  type = "Portable"
  private = true
  ip_version = 4
  capacity = 4
  vlan_id = "${softlayer_virtual_guest.subnetvm1.private_vlan_id}"
  notes = "portable_subnet"
}

resource "softlayer_subnet" "static_subnet" {
  type = "Static"
  private = false
  ip_version = 4
  capacity = 4
  endpoint_ip="${softlayer_virtual_guest.subnetvm1.ipv4_address}"
  notes = "static_subnet"
}

resource "softlayer_subnet" "portable_subnet_v6" {
  type = "Portable"
  private = false
  ip_version = 6
  capacity = 64
  vlan_id = "${softlayer_virtual_guest.subnetvm1.public_vlan_id}"
  notes = "portable_subnet"
}

resource "softlayer_subnet" "static_subnet_v6" {
  type = "Static"
  private = false
  ip_version = 6
  capacity = 64
  endpoint_ip="${softlayer_virtual_guest.subnetvm1.ipv6_address}"
  notes = "static_subnet"
}
`

const testAccCheckSoftLayerSubnetConfig_notes_update = `
resource "softlayer_virtual_guest" "subnetvm1" {
    hostname = "subnetvm1"
    domain = "example.com"
    os_reference_code = "DEBIAN_7_64"
    datacenter = "wdc04"
    network_speed = 100
    hourly_billing = true
    private_network_only = false
    cores = 1
    memory = 1024
    disks = [25]
    local_disk = false
    ipv6_enabled = true
}

resource "softlayer_subnet" "portable_subnet" {
  type = "Portable"
  private = true
  ip_version = 4
  capacity = 4
  vlan_id = "${softlayer_virtual_guest.subnetvm1.private_vlan_id}"
  notes = "portable_subnet_updated"
}

resource "softlayer_subnet" "static_subnet" {
  type = "Static"
  private = false
  ip_version = 4
  capacity = 4
  endpoint_ip="${softlayer_virtual_guest.subnetvm1.ipv4_address}"
  notes = "static_subnet_updated"
}

resource "softlayer_subnet" "portable_subnet_v6" {
  type = "Portable"
  private = false
  ip_version = 6
  capacity = 64
  vlan_id = "${softlayer_virtual_guest.subnetvm1.public_vlan_id}"
  notes = "portable_subnet"
}

resource "softlayer_subnet" "static_subnet_v6" {
  type = "Static"
  private = false
  ip_version = 6
  capacity = 64
  endpoint_ip="${softlayer_virtual_guest.subnetvm1.ipv6_address}"
  notes = "static_subnet"
}
`
