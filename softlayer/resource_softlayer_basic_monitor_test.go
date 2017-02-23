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
)

func TestAccSoftLayerBasicMonitor_Basic(t *testing.T) {
	var basicMonitor datatypes.Network_Monitor_Version1_Query_Host

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerBasicMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSoftLayerBasicMonitorConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerBasicMonitorExists("softlayer_basic_monitor.testacc_foobar", &basicMonitor),
					resource.TestCheckResourceAttrSet(
						"softlayer_basic_monitor.testacc_foobar", "guest_id"),
					resource.TestCheckResourceAttrSet(
						"softlayer_basic_monitor.testacc_foobar", "ip_address"),
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "query_type_id", "1"),
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "response_action_id", "1"),
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "wait_cycles", "5"),
					resource.TestCheckFunc(testAccCheckBasicMonitorNotifiedUsers),
				),
			},

			{
				Config: testAccCheckSoftLayerBasicMonitorConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerBasicMonitorExists("softlayer_basic_monitor.testacc_foobar", &basicMonitor),
					resource.TestCheckResourceAttrSet(
						"softlayer_basic_monitor.testacc_foobar", "guest_id"),
					resource.TestCheckResourceAttrSet(
						"softlayer_basic_monitor.testacc_foobar", "ip_address"),
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "query_type_id", "17"),
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "response_action_id", "2"),
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "wait_cycles", "10"),
					resource.TestCheckFunc(testAccCheckBasicMonitorNotifiedUsers),
				),
			},
		},
	})
}

func testAccCheckSoftLayerBasicMonitorDestroy(s *terraform.State) error {
	service := services.GetNetworkMonitorVersion1QueryHostService(testAccProvider.Meta().(ProviderConfig).SoftLayerSession())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "softlayer_basic_monitor" {
			continue
		}

		basicMonitorId, _ := strconv.Atoi(rs.Primary.ID)

		// Try to find the basic monitor
		_, err := service.Id(basicMonitorId).GetObject()

		if err == nil {
			return errors.New("Basic Monitor still exists")
		}
	}

	return nil
}

func testAccCheckSoftLayerBasicMonitorExists(n string, basicMonitor *datatypes.Network_Monitor_Version1_Query_Host) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Record ID is set")
		}

		basicMonitorId, _ := strconv.Atoi(rs.Primary.ID)

		service := services.GetNetworkMonitorVersion1QueryHostService(testAccProvider.Meta().(ProviderConfig).SoftLayerSession())
		foundBasicMonitor, err := service.Id(basicMonitorId).GetObject()

		if err != nil {
			return err
		}

		if strconv.Itoa(int(*foundBasicMonitor.Id)) != rs.Primary.ID {
			return errors.New("Record not found")
		}

		*basicMonitor = foundBasicMonitor

		return nil
	}
}

func testAccCheckBasicMonitorNotifiedUsers(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "softlayer_basic_monitor" {
			continue
		}

		if n, ok := rs.Primary.Attributes["notified_users.#"]; ok && n != "" && n != "0" {
			return nil
		}

		break
	}

	return errors.New("Basic monitor has no notified users")
}

var testAccCheckSoftLayerBasicMonitorVirtualGuest = `
resource "softlayer_virtual_guest" "terraform-basic-monitor-test" {
    hostname = "terraform-monitor-test"
    domain = "bar.example.com"
    os_reference_code = "DEBIAN_7_64"
    datacenter = "wdc04"
    hourly_billing = true
	private_network_only = false
    cores = 1
    memory = 1024
    local_disk = true
}
`

var testAccCheckSoftLayerBasicMonitorConfig_basic = testAccCheckSoftLayerUserConfig_basic + testAccCheckSoftLayerBasicMonitorVirtualGuest + `
resource "softlayer_basic_monitor" "testacc_foobar" {
    guest_id = "${softlayer_virtual_guest.terraform-basic-monitor-test.id}"
    ip_address = "${softlayer_virtual_guest.terraform-basic-monitor-test.ipv4_address}"
    query_type_id = 1
    response_action_id = 1
    wait_cycles = 5      
    notified_users = ["${softlayer_user.testuser.id}"]
}`

var testAccCheckSoftLayerBasicMonitorConfig_updated = testAccCheckSoftLayerUserConfig_basic + testAccCheckSoftLayerBasicMonitorVirtualGuest + `
resource "softlayer_basic_monitor" "testacc_foobar" {
    guest_id = "${softlayer_virtual_guest.terraform-basic-monitor-test.id}"
    ip_address = "${softlayer_virtual_guest.terraform-basic-monitor-test.ipv4_address}"
    query_type_id = 17
    response_action_id = 2
    wait_cycles = 10
    notified_users = ["${softlayer_user.testuser.id}"]
}`
