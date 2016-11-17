package softlayer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
)

func TestAccSoftLayerVirtualGuest_Basic(t *testing.T) {
	var guest datatypes.Virtual_Guest

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerVirtualGuestDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCheckSoftLayerVirtualGuestConfig_basic,
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerVirtualGuestExists("softlayer_virtual_guest.terraform-acceptance-test-1", &guest),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "hostname", "terraform-test"),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "domain", "bar.example.com"),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "datacenter", "wdc01"),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "network_speed", "10"),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "hourly_billing", "true"),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "private_network_only", "false"),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "cores", "1"),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "memory", "1024"),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "disks.0", "25"),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "disks.1", "10"),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "disks.2", "20"),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "user_metadata", "{\"value\":\"newvalue\"}"),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "local_disk", "false"),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "dedicated_acct_host_only", "true"),
					CheckStringSet(
						"softlayer_virtual_guest.terraform-acceptance-test-1",
						"tags", []string{"collectd"},
					),
				),
			},

			{
				Config:  testAccCheckSoftLayerVirtualGuestConfig_update,
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerVirtualGuestExists("softlayer_virtual_guest.terraform-acceptance-test-1", &guest),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "user_metadata", "updatedData"),
					CheckStringSet(
						"softlayer_virtual_guest.terraform-acceptance-test-1",
						"tags", []string{"mesos-master"},
					),
				),
			},

			{
				Config: testAccCheckSoftLayerVirtualGuestConfig_upgradeMemoryNetworkSpeed,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerVirtualGuestExists("softlayer_virtual_guest.terraform-acceptance-test-1", &guest),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "memory", "2048"),
					resource.TestCheckResourceAttr(
						"softlayer_virtual_guest.terraform-acceptance-test-1", "network_speed", "100"),
				),
			},

			// TODO: currently CPU upgrade test is disabled, due to unexpected behavior of field "dedicated_acct_host_only".
			// TODO: For some reason it is reset by SoftLayer to "false". Daniel Bright reported corresponding issue to SoftLayer team.
			//			{
			//				Config: testAccCheckSoftLayerVirtualGuestConfig_vmUpgradeCPUs,
			//				Check: resource.ComposeTestCheckFunc(
			//					testAccCheckSoftLayerVirtualGuestExists("softlayer_virtual_guest.terraform-acceptance-test-1", &guest),
			//					resource.TestCheckResourceAttr(
			//						"softlayer_virtual_guest.terraform-acceptance-test-1", "cores", "2"),
			//				),
			//			},

		},
	})
}

func TestAccSoftLayerVirtualGuest_BlockDeviceTemplateGroup(t *testing.T) {
	var guest datatypes.Virtual_Guest

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerVirtualGuestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSoftLayerVirtualGuestConfig_blockDeviceTemplateGroup,
				Check: resource.ComposeTestCheckFunc(
					// image_id value is hardcoded. If it's valid then virtual guest will be created well
					testAccCheckSoftLayerVirtualGuestExists("softlayer_virtual_guest.terraform-acceptance-test-BDTGroup", &guest),
				),
			},
		},
	})
}

func TestAccSoftLayerVirtualGuest_postInstallScriptUri(t *testing.T) {
	var guest datatypes.Virtual_Guest

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerVirtualGuestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSoftLayerVirtualGuestConfig_postInstallScriptUri,
				Check: resource.ComposeTestCheckFunc(
					// image_id value is hardcoded. If it's valid then virtual guest will be created well
					testAccCheckSoftLayerVirtualGuestExists("softlayer_virtual_guest.terraform-acceptance-test-pISU", &guest),
				),
			},
		},
	})
}

func testAccCheckSoftLayerVirtualGuestDestroy(s *terraform.State) error {
	service := services.GetVirtualGuestService(testAccProvider.Meta().(ProviderConfig).SoftLayerSession())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "softlayer_virtual_guest" {
			continue
		}

		guestId, _ := strconv.Atoi(rs.Primary.ID)

		// Try to find the guest
		_, err := service.Id(guestId).GetObject()

		// Wait

		if err != nil && !strings.Contains(err.Error(), "404") {
			return fmt.Errorf(
				"Error waiting for virtual guest (%s) to be destroyed: %s",
				rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckSoftLayerVirtualGuestExists(n string, guest *datatypes.Virtual_Guest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No virtual guest ID is set")
		}

		id, err := strconv.Atoi(rs.Primary.ID)

		if err != nil {
			return err
		}

		service := services.GetVirtualGuestService(testAccProvider.Meta().(ProviderConfig).SoftLayerSession())
		retrieveVirtGuest, err := service.Id(id).GetObject()

		if err != nil {
			return err
		}

		fmt.Printf("The ID is %d\n", id)

		if *retrieveVirtGuest.Id != id {
			return errors.New("Virtual guest not found")
		}

		*guest = retrieveVirtGuest

		return nil
	}
}

func CheckStringSet(n string, name string, set []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		values := []string{}
		setLengthKey := fmt.Sprintf("%s.#", name)
		prefix := fmt.Sprintf("%s.", name)
		for k, v := range rs.Primary.Attributes {
			if k != setLengthKey && strings.HasPrefix(k, prefix) {
				values = append(values, v)
			}
		}

		if len(values) == 0 {
			return fmt.Errorf("Could not find %s.%s", n, name)
		}

		for _, s := range set {
			found := false
			for _, v := range values {
				if s == v {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("%s was not found in the set %s", s, name)
			}
		}

		return nil
	}
}

const testAccCheckSoftLayerVirtualGuestConfig_basic = `
resource "softlayer_virtual_guest" "terraform-acceptance-test-1" {
    hostname = "terraform-test"
    domain = "bar.example.com"
    os_reference_code = "DEBIAN_7_64"
    datacenter = "wdc01"
    network_speed = 10
    hourly_billing = true
	private_network_only = false
    cores = 1
    memory = 1024
    disks = [25, 10, 20]
    user_metadata = "{\"value\":\"newvalue\"}"
    tags = ["collectd"]
    dedicated_acct_host_only = true
    local_disk = false
}
`

const testAccCheckSoftLayerVirtualGuestConfig_update = `
resource "softlayer_virtual_guest" "terraform-acceptance-test-1" {
    hostname = "terraform-test"
    domain = "bar.example.com"
    os_reference_code = "DEBIAN_7_64"
    datacenter = "wdc01"
    network_speed = 10
    hourly_billing = true
    cores = 1
    memory = 1024
    disks = [25, 10, 20]
    user_metadata = "updatedData"
    tags = ["mesos-master"]
    dedicated_acct_host_only = true
    local_disk = false
}
`

const testAccCheckSoftLayerVirtualGuestConfig_upgradeMemoryNetworkSpeed = `
resource "softlayer_virtual_guest" "terraform-acceptance-test-1" {
    hostname = "terraform-test"
    domain = "bar.example.com"
    os_reference_code = "DEBIAN_7_64"
    datacenter = "wdc01"
    network_speed = 100
    hourly_billing = true
    cores = 1
    memory = 2048
    disks = [25, 10, 20]
    user_metadata = "updatedData"
    tags = ["mesos-master"]
    dedicated_acct_host_only = true
    local_disk = false
}
`

const testAccCheckSoftLayerVirtualGuestConfig_vmUpgradeCPUs = `
resource "softlayer_virtual_guest" "terraform-acceptance-test-1" {
    hostname = "terraform-test"
    domain = "bar.example.com"
    os_reference_code = "DEBIAN_7_64"
    datacenter = "wdc01"
    network_speed = 100
    hourly_billing = true
    cores = 2
    memory = 2048
    disks = [25, 10, 20]
    user_metadata = "updatedData"
    tags = ["mesos-master"]
    dedicated_acct_host_only = true
    local_disk = false
}
`

const testAccCheckSoftLayerVirtualGuestConfig_postInstallScriptUri = `
resource "softlayer_virtual_guest" "terraform-acceptance-test-pISU" {
    hostname = "terraform-test-pISU"
    domain = "bar.example.com"
    os_reference_code = "DEBIAN_7_64"
    datacenter = "wdc01"
    network_speed = 10
    hourly_billing = true
	private_network_only = false
    cores = 1
    memory = 1024
    disks = [25, 10, 20]
    user_metadata = "{\"value\":\"newvalue\"}"
    dedicated_acct_host_only = true
    local_disk = false
    post_install_script_uri = "https://www.google.com"
}
`

const testAccCheckSoftLayerVirtualGuestConfig_blockDeviceTemplateGroup = `
resource "softlayer_virtual_guest" "terraform-acceptance-test-BDTGroup" {
    hostname = "terraform-test-blockDeviceTemplateGroup"
    domain = "bar.example.com"
    datacenter = "wdc01"
    network_speed = 10
    hourly_billing = false
    cores = 1
    memory = 1024
    local_disk = false
    image_id = 1025457
}
`
