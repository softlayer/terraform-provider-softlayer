package softlayer

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccSoftLayerDnsDomainDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerDnsDomainDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.softlayer_dns_domain.domain_id", "name", "test-domain.com"),
					resource.TestMatchResourceAttr("data.softlayer_dns_domain.domain_id", "id", regexp.MustCompile("^[0-9]+$")),
				),
			},
		},
	})
}

// The datasource to apply
const testAccCheckSoftLayerDnsDomainDataSourceConfig_basic = `

data "softlayer_dns_domain" "domain_id" {
    name = "test-domain.com"
}
`

/*
 * Note: Before running this test, apply the following configuration into the test SoftLayer account:

	resource "softlayer_dns_domain" "domain_1" {
		name = "test-domain.com"
	}

*/
