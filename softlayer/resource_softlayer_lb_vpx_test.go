package softlayer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
)

func TestAccSoftLayerLbVpx_Basic(t *testing.T) {
	var nadc datatypes.Network_Application_Delivery_Controller

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerLbVpxConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerLbVpxExists("softlayer_lb_vpx.testacc_foobar_vpx", &nadc),
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx.testacc_foobar_vpx", "type", "NetScaler VPX"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx.testacc_foobar_vpx", "datacenter", "dal06"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx.testacc_foobar_vpx", "speed", "10"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx.testacc_foobar_vpx", "plan", "Standard"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx.testacc_foobar_vpx", "ip_count", "2"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx.testacc_foobar_vpx", "version", "10.1"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx.testacc_foobar_vpx", "public_vlan_id", "1291213"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx.testacc_foobar_vpx", "private_vlan_id", "1258279"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx.testacc_foobar_vpx", "public_subnet", "184.172.106.152/29"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx.testacc_foobar_vpx", "private_subnet", "10.146.95.64/26"),
					resource.TestCheckResourceAttr(
						"softlayer_lb_vpx.testacc_foobar_vpx", "vip_pool.#", "2"),
				),
			},
		},
	})
}

func testAccCheckSoftLayerLbVpxExists(n string, nadc *datatypes.Network_Application_Delivery_Controller) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		nadcId, _ := strconv.Atoi(rs.Primary.ID)

		service := services.GetNetworkApplicationDeliveryControllerService(testAccProvider.Meta().(ProviderConfig).SoftLayerSession())
		found, err := service.Id(nadcId).GetObject()
		if err != nil {
			return err
		}

		if strconv.Itoa(int(*found.Id)) != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		*nadc = found

		return nil
	}
}

const testAccCheckSoftLayerLbVpxConfig_basic = `
resource "softlayer_lb_vpx" "testacc_foobar_vpx" {
    datacenter = "dal06"
    speed = 10
    version = "10.1"
    plan = "Standard"
    ip_count = 2
    public_vlan_id = 1291213
    private_vlan_id = 1258279
    public_subnet = "184.172.106.152/29"
    private_subnet = "10.146.95.64/26"
}`
