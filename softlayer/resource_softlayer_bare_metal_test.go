package softlayer

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

func TestAccSoftLayerBareMetal_Basic(t *testing.T) {
	var bareMetal datatypes.Hardware

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerBareMetalDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCheckSoftLayerBareMetalConfig_basic,
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerBareMetalExists("softlayer_bare_metal.terraform-acceptance-test-1", &bareMetal),
					resource.TestCheckResourceAttr(
						"softlayer_bare_metal.terraform-acceptance-test-1", "hostname", "terraform-test"),
					resource.TestCheckResourceAttr(
						"softlayer_bare_metal.terraform-acceptance-test-1", "domain", "bar.example.com"),
					resource.TestCheckResourceAttr(
						"softlayer_bare_metal.terraform-acceptance-test-1", "datacenter", "dal01"),
					resource.TestCheckResourceAttr(
						"softlayer_bare_metal.terraform-acceptance-test-1", "network_speed", "100"),
					resource.TestCheckResourceAttr(
						"softlayer_bare_metal.terraform-acceptance-test-1", "hourly_billing", "true"),
					resource.TestCheckResourceAttr(
						"softlayer_bare_metal.terraform-acceptance-test-1", "private_network_only", "false"),
					resource.TestCheckResourceAttr(
						"softlayer_bare_metal.terraform-acceptance-test-1", "user_metadata", "{\"value\":\"newvalue\"}"),
					resource.TestCheckResourceAttr(
						"softlayer_bare_metal.terraform-acceptance-test-1", "fixed_config_preset", "S1270_8GB_2X1TBSATA_NORAID"),
				),
			},
		},
	})
}

func testAccCheckSoftLayerBareMetalDestroy(s *terraform.State) error {
	service := services.GetHardwareService(testAccProvider.Meta().(*session.Session))

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "softlayer_bare_metal" {
			continue
		}

		id, _ := strconv.Atoi(rs.Primary.ID)

		// Try to find the bare metal
		_, err := service.Id(id).GetObject()

		// Wait
		if err != nil {
			if apiErr, ok := err.(sl.Error); !ok || apiErr.StatusCode != 404 {
				return fmt.Errorf(
					"Error waiting for bare metal (%d) to be destroyed: %s",
					id, err)
			}
		}
	}

	return nil
}

func testAccCheckSoftLayerBareMetalExists(n string, bareMetal *datatypes.Hardware) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No bare metal ID is set")
		}

		id, err := strconv.Atoi(rs.Primary.ID)

		if err != nil {
			return err
		}

		service := services.GetHardwareService(testAccProvider.Meta().(*session.Session))
		bm, err := service.Id(id).GetObject()
		if err != nil {
			return err
		}

		fmt.Printf("The ID is %d", *bm.Id)

		if *bm.Id != id {
			return errors.New("Bare metal not found")
		}

		*bareMetal = bm

		return nil
	}
}

const testAccCheckSoftLayerBareMetalConfig_basic = `
resource "softlayer_bare_metal" "terraform-acceptance-test-1" {
    hostname = "terraform-test"
    domain = "bar.example.com"
    os_reference_code = "UBUNTU_16_64"
    datacenter = "ams01"
    network_speed = 100
    hourly_billing = true
    private_network_only = false
    user_metadata = "{\"value\":\"newvalue\"}"
    fixed_config_preset = "S1270_8GB_2X1TBSATA_NORAID"
}
`
