package softlayer

import (
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

func TestAccSoftLayerVlan_Basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerVlanConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"softlayer_vlan.test_vlan", "name", "test_vlan"),
					resource.TestCheckResourceAttr(
						"softlayer_vlan.test_vlan", "datacenter", "lon02"),
					resource.TestCheckResourceAttr(
						"softlayer_vlan.test_vlan", "type", "PUBLIC"),
					resource.TestCheckResourceAttr(
						"softlayer_vlan.test_vlan", "softlayer_managed", "false"),
					resource.TestCheckResourceAttr(
						"softlayer_vlan.test_vlan", "primary_router_hostname", "fcr01a.lon02"),
					resource.TestCheckResourceAttr(
						"softlayer_vlan.test_vlan", "primary_subnet_size", "8"),
				),
			},

			resource.TestStep{
				Config: testAccCheckSoftLayerVlanConfig_name_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"softlayer_vlan.test_vlan", "name", "test_vlan_update"),
				),
			},
		},
	})
}

const testAccCheckSoftLayerVlanConfig_basic = `
resource "softlayer_vlan" "test_vlan" {
   name = "test_vlan"
   datacenter = "lon02"
   type = "PUBLIC"
   primary_subnet_size = 8
   primary_router_hostname = "fcr01a.lon02"
}`

const testAccCheckSoftLayerVlanConfig_name_update = `
resource "softlayer_vlan" "test_vlan" {
   name = "test_vlan_update"
   datacenter = "lon02"
   type = "PUBLIC"
   primary_subnet_size = 8
   primary_router_hostname = "fcr01a.lon02"
}`
