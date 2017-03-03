package softlayer

import (
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

func TestAccSoftLayerFwHardwareDedicatedRules_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerFwHardwareDedicatedRules_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.0.action", "deny"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.0.src_ip_address", "0.0.0.0"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.0.dst_ip_address", "any"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.0.dst_port_range_start", "1"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.0.dst_port_range_end", "65535"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.0.notes", "Deny all"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.0.protocol", "tcp"),

					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.1.action", "permit"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.1.src_ip_address", "0.0.0.0"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.1.dst_ip_address", "any"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.1.dst_port_range_start", "22"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.1.dst_port_range_end", "22"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.1.notes", "Allow SSH"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.1.protocol", "tcp"),

					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.2.action", "permit"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.2.src_ip_address",
						"0000:0000:0000:0000:0000:0000:0000:0000"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.2.dst_ip_address", "any"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.2.dst_port_range_start", "22"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.2.dst_port_range_end", "22"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.2.notes", "Allow SSH"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.2.protocol", "tcp"),
				),
			},
			resource.TestStep{
				Config: testAccCheckSoftLayerFwHardwareDedicatedRules_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.0.action", "permit"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.0.src_ip_address", "10.1.1.0"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.0.dst_port_range_start", "80"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.0.dst_port_range_end", "80"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.0.notes", "Permit from 10.1.1.0"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.0.protocol", "udp"),

					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.1.action", "deny"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.1.src_ip_address", "2401:c900:1501:0032:0000:0000:0000:0000"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.1.dst_port_range_start", "80"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.1.dst_port_range_end", "80"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.1.notes", "Deny for IPv6"),
					resource.TestCheckResourceAttr(
						"softlayer_fw_hardware_dedicated_rules.rules", "rules.1.protocol", "udp"),
				),
			},
		},
	})
}

const testAccCheckSoftLayerFwHardwareDedicatedRules_basic = `
resource "softlayer_virtual_guest" "fwvm2" {
    hostname = "fwvm2"
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

resource "softlayer_fw_hardware_dedicated" "accfw2" {
  ha_enabled = false
  public_vlan_id = "${softlayer_virtual_guest.fwvm2.public_vlan_id}"
}

resource "softlayer_fw_hardware_dedicated_rules" "rules" {
 firewall_id = "${softlayer_fw_hardware_dedicated.accfw2.id}"
 rules = {
      "action" = "deny"
      "src_ip_address"= "0.0.0.0"
      "src_ip_cidr"= 0
      "dst_ip_address"= "any"
      "dst_ip_cidr"= 32
      "dst_port_range_start"= 1
      "dst_port_range_end"= 65535
      "notes"= "Deny all"
      "protocol"= "tcp"
 }
 rules = {
      "action" = "permit"
      "src_ip_address"= "0.0.0.0"
      "src_ip_cidr"= 0
      "dst_ip_address"= "any"
      "dst_ip_cidr"= 32
      "dst_port_range_start"= 22
      "dst_port_range_end"= 22
      "notes"= "Allow SSH"
      "protocol"= "tcp"
 }
 rules = {
      "action" = "permit"
      "src_ip_address"= "0::"
      "src_ip_cidr"= 0
      "dst_ip_address"= "any"
      "dst_ip_cidr"= 128
      "dst_port_range_start"= 22
      "dst_port_range_end"= 22
      "notes"= "Allow SSH"
      "protocol"= "tcp"
 }
}

`

const testAccCheckSoftLayerFwHardwareDedicatedRules_update = `
resource "softlayer_virtual_guest" "fwvm2" {
    hostname = "fwvm2"
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

resource "softlayer_fw_hardware_dedicated" "accfw2" {
  ha_enabled = false
  public_vlan_id = "${softlayer_virtual_guest.fwvm2.public_vlan_id}"
}

resource "softlayer_fw_hardware_dedicated_rules" "rules" {
 firewall_id = "${softlayer_fw_hardware_dedicated.accfw2.id}"
 rules = {
      "action" = "permit"
      "src_ip_address"= "10.1.1.0"
      "src_ip_cidr"= 24
      "dst_ip_address"= "any"
      "dst_ip_cidr"= 32
      "dst_port_range_start"= 80
      "dst_port_range_end"= 80
      "notes"= "Permit from 10.1.1.0"
      "protocol"= "udp"
 }
 rules = {
      "action" = "deny"
      "src_ip_address"= "2401:c900:1501:0032:0000:0000:0000:0000"
      "src_ip_cidr"= 64
      "dst_ip_address"= "any"
      "dst_ip_cidr"= 128
      "dst_port_range_start"= 80
      "dst_port_range_end"= 80
      "notes"= "Deny for IPv6"
      "protocol"= "udp"
 }
}
`
