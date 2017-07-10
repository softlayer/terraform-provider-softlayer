package softlayer

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccSoftLayerQuoteBareMetalDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSoftLayerQuoteBareMetalDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.softlayer_quote_bare_metal.test_quote_bm",
						"package_key_name",
						"4U_DUAL_E52600_36_DRIVES",
					),
					resource.TestCheckResourceAttr(
						"data.softlayer_quote_bare_metal.test_quote_bm",
						"process_key_name",
						"INTEL_XEON_2650_2_00",
					),
					resource.TestCheckResourceAttr(
						"data.softlayer_quote_bare_metal.test_quote_bm",
						"datacenter",
						"dal06",
					),
				),
			},
		},
	})
}

// The datasource to apply
const testAccCheckSoftLayerQuoteBareMetalDataSourceConfig_basic = `
data "softlayer_quote_bare_metal" "test_quote_bm" {
    name = "test"
}
`
