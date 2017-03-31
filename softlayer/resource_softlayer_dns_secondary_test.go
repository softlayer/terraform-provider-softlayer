package softlayer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/softlayer/softlayer-go/services"
)

func TestAccSoftLayerDnsSecondary_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSoftLayerDnsSecondaryConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerDnsSecondaryZoneExists("softlayer_dns_secondary.dns-secondary-zone-1"),
					resource.TestCheckResourceAttr(
						"softlayer_dns_secondary.dns-secondary-zone-1", "zoneName", "new-secondary-zone1.com"),
					resource.TestCheckResourceAttr(
						"softlayer_dns_secondary.dns-secondary-zone-1", "transferFrequency", "10"),
					resource.TestCheckResourceAttr(
						"softlayer_dns_secondary.dns-secondary-zone-1", "masterIpAddress", "172.16.0.1"),
				),
			},
			{
				Config: testAccCheckSoftLayerDnsSecondaryConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"softlayer_dns_secondary.dns-secondary-zone-1", "transferFrequency", "15"),
					resource.TestCheckResourceAttr(
						"softlayer_dns_secondary.dns-secondary-zone-1", "masterIpAddress", "172.16.0.2"),
				),
			},
		},
	})
}

func testAccCheckSoftLayerDnsSecondaryZoneExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		dnsId, _ := strconv.Atoi(rs.Primary.ID)

		service := services.GetDnsSecondaryService(testAccProvider.Meta().(ProviderConfig).SoftLayerSession())
		foundSecondaryZone, err := service.Id(dnsId).GetObject()

		if err != nil {
			return err
		}

		if strconv.Itoa(int(*foundSecondaryZone.Id)) != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		return nil
	}
}

const testAccCheckSoftLayerDnsSecondaryConfig_basic = `
resource "softlayer_dns_secondary" "dns-secondary-zone-1" {
    zoneName = "new-secondary-zone1.com"
    transferFrequency = 10
    masterIpAddress = "172.16.0.1"
}
`

const testAccCheckSoftLayerDnsSecondaryConfig_updated = `
resource "softlayer_dns_secondary" "dns-secondary-zone-1" {
    zoneName = "new-secondary-zone1.com"
    transferFrequency = 15
    masterIpAddress = "172.16.0.2"
}
`
