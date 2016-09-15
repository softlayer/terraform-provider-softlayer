package softlayer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
)

func TestAccSoftLayerNetworkVlan_Basic(t *testing.T) {
	var vlan datatypes.Network_Vlan

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerNetworkVlanDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerNetworkVlanConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerNetworkVlanExists("softlayer_network_vlan.test-network-vlan", &vlan),
					testAccCheckSoftLayerNetworkVlanAttributes(&vlan),
					resource.TestCheckResourceAttr(
						"softlayer_network_vlan.test-network-vlan", "name", "test-network-vlan"),
					resource.TestCheckResourceAttr(
						"softlayer_network_vlan.test-network-vlan", "datacenter", "sjc03"),
					resource.TestCheckResourceAttr(
						"softlayer_network_vlan.test-network-vlan", "type", "private"),
					resource.TestCheckResourceAttr(
						"softlayer_network_vlan.test-network-vlan", "note", "vlan for test"),
					resource.TestCheckResourceAttr(
						"softlayer_network_vlan.test-network-vlan", "primary_router_hostname", "bcr01a.sjc03"),
					resource.TestCheckResourceAttr(
						"softlayer_network_vlan.test-network-vlan", "vlan_number", 1424),
				),
			},
		},
	})
}

func testAccCheckSoftLayerNetworkVlanDestroy(s *terraform.State) error {
	service := services.GetNetworkVlanService(testAccProvider.Meta().(*session.Session))

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "softlayer_network_vlan" {
			continue
		}

		vlanId, _ := strconv.Atoi(rs.Primary.ID)

		// Try to find the network vlan
		_, err := service.Id(vlanId).GetObject()

		if err == nil {
			return fmt.Errorf("Network Vlan still exists")
		}
	}

	return nil
}

func testAccCheckSoftLayerNetworkVlanAttributes(vlan *datatypes.Network_Vlan) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *vlan.Name != "test-network-vlan" {
			return fmt.Errorf("Bad name: %s", *vlan.Name)
		}

		if *vlan.Note != "vlan for test" {
			return fmt.Errorf("Bad note: %s", *vlan.Note)
		}
		if *vlan.PrimaryRouter.Hostname != "bcr01a.sjc03" {
			return fmt.Errorf("Bad primary router hostname: %s", *vlan.PrimaryRouter.Hostname)
		}
		if *vlan.VlanNumber != 1424 {
			return fmt.Errorf("Bad vlan number: %s", *vlan.VlanNumber)
		}

		return nil
	}
}

func testAccCheckSoftLayerNetworkVlanExists(n string, vlan *datatypes.Network_Vlan) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		vlanId, _ := strconv.Atoi(rs.Primary.ID)

		service := services.GetNetworkVlanService(testAccProvider.Meta().(*session.Session))
		foundVlan, err := service.Id(vlanId).GetObject()

		if err != nil {
			return err
		}

		if strconv.Itoa(int(*foundVlan.Id)) != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		*vlan = foundVlan

		return nil
	}
}

const testAccCheckSoftLayerNetworkVlanConfig_basic = `
resource "softlayer_network_vlan" "test-network-vlan" {
	name = "test-vlan-1"
	datacenter = "sjc03"
	type = "private"
	note = "vlan for test"
	primary_router_hostname = "bcr01a.sjc03"
	vlan_number = 1424
}`
