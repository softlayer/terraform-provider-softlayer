package softlayer

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccSoftLayerVlanDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSoftLayerVlanDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.softlayer_vlan.tfacc_vlan",
						"number",
						"870",
					),
					resource.TestCheckResourceAttr(
						"data.softlayer_vlan.tfacc_vlan",
						"router_hostname",
						"bcr05.dal01",
					),
					resource.TestMatchResourceAttr(
						"data.softlayer_vlan.tfacc_vlan",
						"id",
						regexp.MustCompile("^[0-9]+$"),
					),
				),
			},
		},
	})
}

// The datasource to apply
const testAccCheckSoftLayerVlanDataSourceConfig_basic = `
data "softlayer_vlan" "tfacc_vlan" {
    number = 870
    router_hostname = "bcr05.dal01"
}
`
