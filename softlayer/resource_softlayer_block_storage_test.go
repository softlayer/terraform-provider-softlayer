package softlayer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/softlayer/softlayer-go/services"
)

func TestAccSoftLayerBlockStorage_Basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerBlockStorageConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					// Endurance Storage
					testAccCheckSoftLayerBlockStorageExists("softlayer_block_storage.bs_endurance"),
					resource.TestCheckResourceAttr(
						"softlayer_block_storage.bs_endurance", "type", "Endurance"),
					resource.TestCheckResourceAttr(
						"softlayer_block_storage.bs_endurance", "capacity", "20"),
					resource.TestCheckResourceAttr(
						"softlayer_block_storage.bs_endurance", "iops", "0.25"),
					resource.TestCheckResourceAttr(
						"softlayer_block_storage.bs_endurance", "snapshot_capacity", "10"),
					resource.TestCheckResourceAttr(
						"softlayer_block_storage.bs_endurance", "os_format_type", "Linux"),
					testAccCheckSoftLayerResources("softlayer_block_storage.bs_endurance", "datacenter",
						"softlayer_virtual_guest.storagevm2", "datacenter"),
					// Performance Storage
					testAccCheckSoftLayerBlockStorageExists("softlayer_block_storage.bs_performance"),
					resource.TestCheckResourceAttr(
						"softlayer_block_storage.bs_performance", "type", "Performance"),
					resource.TestCheckResourceAttr(
						"softlayer_block_storage.bs_performance", "capacity", "20"),
					resource.TestCheckResourceAttr(
						"softlayer_block_storage.bs_performance", "iops", "100"),
					resource.TestCheckResourceAttr(
						"softlayer_block_storage.bs_endurance", "os_format_type", "Linux"),
					testAccCheckSoftLayerResources("softlayer_block_storage.bs_performance", "datacenter",
						"softlayer_virtual_guest.storagevm2", "datacenter"),
				),
			},

			resource.TestStep{
				Config: testAccCheckSoftLayerBlockStorageConfig_update,
				Check: resource.ComposeTestCheckFunc(
					// Endurance Storage
					resource.TestCheckResourceAttr("softlayer_block_storage.bs_endurance", "allowed_virtual_guest_ids.#", "1"),
					resource.TestCheckResourceAttr("softlayer_block_storage.bs_endurance", "allowed_ip_addresses.#", "1"),
					// Performance Storage
					resource.TestCheckResourceAttr("softlayer_block_storage.bs_performance", "allowed_virtual_guest_ids.#", "1"),
					resource.TestCheckResourceAttr("softlayer_block_storage.bs_performance", "allowed_ip_addresses.#", "1"),
				),
			},
		},
	})
}

func testAccCheckSoftLayerBlockStorageExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		storageId, _ := strconv.Atoi(rs.Primary.ID)

		service := services.GetNetworkStorageService(testAccProvider.Meta().(ProviderConfig).SoftLayerSession())
		foundStorage, err := service.Id(storageId).GetObject()

		if err != nil {
			return err
		}

		if strconv.Itoa(int(*foundStorage.Id)) != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		return nil
	}
}

const testAccCheckSoftLayerBlockStorageConfig_basic = `
resource "softlayer_virtual_guest" "storagevm2" {
    hostname = "storagevm2"
    domain = "example.com"
    os_reference_code = "DEBIAN_7_64"
    datacenter = "dal06"
    network_speed = 100
    hourly_billing = true
    private_network_only = false
    cores = 1
    memory = 1024
    disks = [25]
    local_disk = false
}

resource "softlayer_block_storage" "bs_endurance" {
        type = "Endurance"
        datacenter = "${softlayer_virtual_guest.storagevm2.datacenter}"
        capacity = 20
        iops = 0.25
        snapshot_capacity = 10
        os_format_type = "Linux"
}

resource "softlayer_block_storage" "bs_performance" {
        type = "Performance"
        datacenter = "${softlayer_virtual_guest.storagevm2.datacenter}"
        capacity = 20
        iops = 100
        os_format_type = "Linux"
}
`
const testAccCheckSoftLayerBlockStorageConfig_update = `
resource "softlayer_virtual_guest" "storagevm2" {
    hostname = "storagevm2"
    domain = "example.com"
    os_reference_code = "DEBIAN_7_64"
    datacenter = "dal06"
    network_speed = 100
    hourly_billing = true
    private_network_only = false
    cores = 1
    memory = 1024
    disks = [25]
    local_disk = false
}

resource "softlayer_block_storage" "bs_endurance" {
        type = "Endurance"
        datacenter = "${softlayer_virtual_guest.storagevm2.datacenter}"
        capacity = 20
        iops = 0.25
        os_format_type = "Linux"
        allowed_virtual_guest_ids = [ "${softlayer_virtual_guest.storagevm2.id}" ]
        allowed_ip_addresses = [ "${softlayer_virtual_guest.storagevm2.ipv4_address_private}" ]
        snapshot_capacity = 10
}

resource "softlayer_block_storage" "bs_performance" {
        type = "Performance"
        datacenter = "${softlayer_virtual_guest.storagevm2.datacenter}"
        capacity = 20
        iops = 100
        os_format_type = "Linux"
        allowed_virtual_guest_ids = [ "${softlayer_virtual_guest.storagevm2.id}" ]
        allowed_ip_addresses = [ "${softlayer_virtual_guest.storagevm2.ipv4_address_private}" ]
}
`
