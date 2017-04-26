package softlayer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/softlayer/softlayer-go/services"
)

func TestAccSoftLayerFileStorage_Basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerFileStorageConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					// Endurance Storage
					testAccCheckSoftLayerFileStorageExists("softlayer_file_storage.fs_endurance"),
					resource.TestCheckResourceAttr(
						"softlayer_file_storage.fs_endurance", "type", "Endurance"),
					resource.TestCheckResourceAttr(
						"softlayer_file_storage.fs_endurance", "capacity", "20"),
					resource.TestCheckResourceAttr(
						"softlayer_file_storage.fs_endurance", "iops", "0.25"),
					resource.TestCheckResourceAttr(
						"softlayer_file_storage.fs_endurance", "snapshot_capacity", "10"),
					testAccCheckSoftLayerResources("softlayer_file_storage.fs_endurance", "datacenter",
						"softlayer_virtual_guest.storagevm1", "datacenter"),
					resource.TestCheckResourceAttr("softlayer_file_storage.fs_endurance", "notes", "endurance notes"),
					resource.TestCheckResourceAttrSet("softlayer_file_storage.fs_endurance", "mountpoint"),
					// Performance Storage
					testAccCheckSoftLayerFileStorageExists("softlayer_file_storage.fs_performance"),
					resource.TestCheckResourceAttr(
						"softlayer_file_storage.fs_performance", "type", "Performance"),
					resource.TestCheckResourceAttr(
						"softlayer_file_storage.fs_performance", "capacity", "20"),
					resource.TestCheckResourceAttr(
						"softlayer_file_storage.fs_performance", "iops", "100"),
					testAccCheckSoftLayerResources("softlayer_file_storage.fs_performance", "datacenter",
						"softlayer_virtual_guest.storagevm1", "datacenter"),
					resource.TestCheckResourceAttr("softlayer_file_storage.fs_performance", "notes", "performance notes"),
					resource.TestCheckResourceAttrSet("softlayer_file_storage.fs_performance", "mountpoint"),
				),
			},

			resource.TestStep{
				Config: testAccCheckSoftLayerFileStorageConfig_update,
				Check: resource.ComposeTestCheckFunc(
					// Endurance Storage
					resource.TestCheckResourceAttr("softlayer_file_storage.fs_endurance", "allowed_virtual_guest_ids.#", "1"),
					resource.TestCheckResourceAttr("softlayer_file_storage.fs_endurance", "allowed_subnets.#", "1"),
					resource.TestCheckResourceAttr("softlayer_file_storage.fs_endurance", "allowed_ip_addresses.#", "1"),
					resource.TestCheckResourceAttr("softlayer_file_storage.fs_endurance", "notes", "updated endurance notes"),
					// Performance Storage
					resource.TestCheckResourceAttr("softlayer_file_storage.fs_performance", "allowed_virtual_guest_ids.#", "1"),
					resource.TestCheckResourceAttr("softlayer_file_storage.fs_performance", "allowed_subnets.#", "1"),
					resource.TestCheckResourceAttr("softlayer_file_storage.fs_performance", "allowed_ip_addresses.#", "1"),
					resource.TestCheckResourceAttr("softlayer_file_storage.fs_performance", "notes", ""),
				),
			},
		},
	})
}

func testAccCheckSoftLayerFileStorageExists(n string) resource.TestCheckFunc {
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

const testAccCheckSoftLayerFileStorageConfig_basic = `
resource "softlayer_virtual_guest" "storagevm1" {
    hostname = "storagevm1"
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

resource "softlayer_file_storage" "fs_endurance" {
        type = "Endurance"
        datacenter = "${softlayer_virtual_guest.storagevm1.datacenter}"
        capacity = 20
        iops = 0.25
        snapshot_capacity = 10
        notes = "endurance notes"
}

resource "softlayer_file_storage" "fs_performance" {
        type = "Performance"
        datacenter = "${softlayer_virtual_guest.storagevm1.datacenter}"
        capacity = 20
        iops = 100
        notes = "performance notes"
}
`
const testAccCheckSoftLayerFileStorageConfig_update = `
resource "softlayer_virtual_guest" "storagevm1" {
    hostname = "storagevm1"
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

resource "softlayer_file_storage" "fs_endurance" {
        type = "Endurance"
        datacenter = "${softlayer_virtual_guest.storagevm1.datacenter}"
        capacity = 20
        iops = 0.25
        allowed_virtual_guest_ids = [ "${softlayer_virtual_guest.storagevm1.id}" ]
        allowed_subnets = [ "${softlayer_virtual_guest.storagevm1.private_subnet}" ]
        allowed_ip_addresses = [ "${softlayer_virtual_guest.storagevm1.ipv4_address_private}" ]
        snapshot_capacity = 10
        notes = "updated endurance notes"
}

resource "softlayer_file_storage" "fs_performance" {
        type = "Performance"
        datacenter = "${softlayer_virtual_guest.storagevm1.datacenter}"
        capacity = 20
        iops = 100
        allowed_virtual_guest_ids = [ "${softlayer_virtual_guest.storagevm1.id}" ]
        allowed_subnets = [ "${softlayer_virtual_guest.storagevm1.private_subnet}" ]
        allowed_ip_addresses = [ "${softlayer_virtual_guest.storagevm1.ipv4_address_private}" ]
}
`
