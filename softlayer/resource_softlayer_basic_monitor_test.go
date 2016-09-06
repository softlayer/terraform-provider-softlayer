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
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "guest_id", "22274327"),
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "ip_address", "169.54.168.102"),
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "query_type_id", "1"),
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "response_action_id", "1"),
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "wait_cycles", "5"),
					testAccCheckSoftLayerBasicMonitorContainsUsers("softlayer_basic_monitor.testacc_foobar", 460547),
				),
			},

			{
				Config: testAccCheckSoftLayerBasicMonitorConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerBasicMonitorExists("softlayer_basic_monitor.testacc_foobar", &basicMonitor),
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "guest_id", "22274327"),
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "ip_address", "169.54.168.102"),
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "query_type_id", "17"),
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "response_action_id", "2"),
					resource.TestCheckResourceAttr(
						"softlayer_basic_monitor.testacc_foobar", "wait_cycles", "10"),
				),
			},
		},
	})
}

func testAccCheckSoftLayerBasicMonitorDestroy(s *terraform.State) error {
	service := services.GetNetworkMonitorVersion1QueryHostService(testAccProvider.Meta().(*session.Session))

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

		service := services.GetNetworkMonitorVersion1QueryHostService(testAccProvider.Meta().(*session.Session))
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

func testAccCheckSoftLayerBasicMonitorContainsUsers(n string, userId int) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Record ID is set")
		}

		basicMonitorId, _ := strconv.Atoi(rs.Primary.ID)

		service := services.GetNetworkMonitorVersion1QueryHostService(testAccProvider.Meta().(*session.Session))
		basicMonitor, err := service.Id(basicMonitorId).GetObject()

		if err != nil {
			return err
		}

		if strconv.Itoa(int(*basicMonitor.Id)) != rs.Primary.ID {
			return errors.New("Record not found")
		}

		notificationLinks, err := services.GetVirtualGuestService(testAccProvider.Meta().(*session.Session)).Mask("userId").Id(*basicMonitor.GuestId).GetMonitoringUserNotification()

		if notificationLinks == nil {
			return errors.New("Cannot get a ust list")
		}

		found := false

		for _, notifiedUser := range notificationLinks {
			if *notifiedUser.UserId == userId {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("UserId %d is not found in notifiedUsers", userId)

		}

		return nil
	}
}

const testAccCheckSoftLayerBasicMonitorConfig_basic = `

resource "softlayer_basic_monitor" "testacc_foobar" {
    guest_id = 22274327
    ip_address = "169.54.168.102"
    query_type_id = 1
    response_action_id = 1
    wait_cycles = 5      
    notified_users = [460547]
}`

const testAccCheckSoftLayerBasicMonitorConfig_updated = `
resource "softlayer_basic_monitor" "testacc_foobar" {
    guest_id = 22274327
    ip_address = "169.54.168.102"
    query_type_id = 17
    response_action_id = 2
    wait_cycles = 10
    notified_users = [460547]
}`
