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
	"github.com/softlayer/softlayer-go/sl"
)

func TestAccSoftLayerBareMetalQuote_Basic(t *testing.T) {
	var bareMetal datatypes.Hardware

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerBareMetalQuoteDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCheckSoftLayerBareMetalQuoteConfig_basic,
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerBareMetalQuoteExists("softlayer_bare_metal_quote.terraform-acceptance-test-1", &bareMetal),
					resource.TestCheckResourceAttr(
						"softlayer_bare_metal_quote.terraform-acceptance-test-1", "hostname", "terraform-test"),
					resource.TestCheckResourceAttr(
						"softlayer_bare_metal_quote.terraform-acceptance-test-1", "domain", "bar.example.com"),
					resource.TestCheckResourceAttr(
						"softlayer_bare_metal_quote.terraform-acceptance-test-1", "user_metadata", "{\"value\":\"newvalue\"}"),
					resource.TestCheckResourceAttr(
						"softlayer_bare_metal_quote.terraform-acceptance-test-1", "quote_id", "2179879"),
					CheckStringSet(
						"softlayer_bare_metal_quote.terraform-acceptance-test-1",
						"tags", []string{"collectd"},
					),
				),
			},

			{
				Config:  testAccCheckSoftLayerBareMetalQuoteConfig_update,
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerBareMetalQuoteExists("softlayer_bare_metal_quote.terraform-acceptance-test-1", &bareMetal),
					CheckStringSet(
						"softlayer_bare_metal_quote.terraform-acceptance-test-1",
						"tags", []string{"mesos-master"},
					),
				),
			},
		},
	})
}

func testAccCheckSoftLayerBareMetalQuoteDestroy(s *terraform.State) error {
	service := services.GetHardwareService(testAccProvider.Meta().(ProviderConfig).SoftLayerSession())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "softlayer_bare_metal_quote" {
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

func testAccCheckSoftLayerBareMetalQuoteExists(n string, bareMetal *datatypes.Hardware) resource.TestCheckFunc {
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

		service := services.GetHardwareService(testAccProvider.Meta().(ProviderConfig).SoftLayerSession())
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

const testAccCheckSoftLayerBareMetalQuoteConfig_basic = `
resource "softlayer_bare_metal_quote" "terraform-acceptance-test-1" {
    hostname = "terraform-test"
    domain = "bar.example.com"
    user_metadata = "{\"value\":\"newvalue\"}"
    quote_id = 2179879
    tags = ["collectd"]
}
`

const testAccCheckSoftLayerBareMetalQuoteConfig_update = `
resource "softlayer_bare_metal_quote" "terraform-acceptance-test-1" {
    hostname = "terraform-test"
    domain = "bar.example.com"
    user_metadata = "{\"value\":\"newvalue\"}"
    quote_id = 2179879
    tags = ["mesos-master"]
}
`
