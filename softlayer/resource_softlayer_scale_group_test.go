package softlayer

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"strings"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
)

func TestAccSoftLayerScaleGroup_Basic(t *testing.T) {
	var scalegroup datatypes.Scale_Group

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerScaleGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerScaleGroupConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerScaleGroupExists("softlayer_scale_group.sample-http-cluster", &scalegroup),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "name", "sample-http-cluster"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "regional_group", "as-sgp-central-1"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "cooldown", "30"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "minimum_member_count", "1"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "maximum_member_count", "10"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "termination_policy", "CLOSEST_TO_NEXT_CHARGE"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "port", "8080"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "health_check.type", "HTTP"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.name", "test-VM"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.domain", "example.com"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.cpu", "1"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.ram", "4096"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.public_network_speed", "1000"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.hourly_billing", "true"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.image", "DEBIAN_7_64"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.local_disk", "false"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.disks.0", "25"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.disks.1", "100"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.datacenter", "sng01"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.post_install_script_uri", ""),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.user_data", "#!/bin/bash"),
					testAccCheckSoftLayerScaleGroupContainsNetworkVlan(&scalegroup, 1928, "bcr02a.sng01"),
				),
			},

			resource.TestStep{
				Config: testAccCheckSoftLayerScaleGroupConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerScaleGroupExists("softlayer_scale_group.sample-http-cluster", &scalegroup),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "name", "changed_name"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "regional_group", "as-sgp-central-1"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "minimum_member_count", "2"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "maximum_member_count", "12"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "termination_policy", "NEWEST"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "cooldown", "35"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "port", "9090"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "health_check.type", "HTTP-CUSTOM"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.name", "example-VM"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.domain", "test.com"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.cpu", "2"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.ram", "8192"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.public_network_speed", "100"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.image", "CENTOS_7_64"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.datacenter", "sng01"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_group.sample-http-cluster", "virtual_guest_member_template.0.post_install_script_uri", "http://localhost/index.html"),
				),
			},
		},
	})
}

func testAccCheckSoftLayerScaleGroupDestroy(s *terraform.State) error {
	service := services.GetScaleGroupService(testAccProvider.Meta().(*session.Session))

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "softlayer_scale_group" {
			continue
		}

		scalegroupId, _ := strconv.Atoi(rs.Primary.ID)

		// Try to find the key
		_, err := service.Id(scalegroupId).Mask("id").GetObject()

		if err != nil && !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("Error waiting for Auto Scale (%s) to be destroyed: %s", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckSoftLayerScaleGroupContainsNetworkVlan(scaleGroup *datatypes.Scale_Group, vlanNumber int, primaryRouterHostname string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		found := false

		for _, scaleVlan := range scaleGroup.NetworkVlans {
			vlan := *scaleVlan.NetworkVlan

			if *vlan.VlanNumber == vlanNumber && *vlan.PrimaryRouter.Hostname == primaryRouterHostname {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf(
				"Vlan number %d with router hostname %s not found in scale group",
				vlanNumber,
				primaryRouterHostname)
		}

		return nil
	}
}

func testAccCheckSoftLayerScaleGroupExists(n string, scalegroup *datatypes.Scale_Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Record ID is set")
		}

		scalegroupId, _ := strconv.Atoi(rs.Primary.ID)

		service := services.GetScaleGroupService(testAccProvider.Meta().(*session.Session))
		foundScaleGroup, err := service.Id(scalegroupId).Mask(strings.Join(SoftLayerScaleGroupObjectMask, ",")).GetObject()

		if err != nil {
			return err
		}

		if strconv.Itoa(int(*foundScaleGroup.Id)) != rs.Primary.ID {
			return fmt.Errorf("Record %s not found", rs.Primary.ID)
		}

		*scalegroup = foundScaleGroup

		return nil
	}
}

const testAccCheckSoftLayerScaleGroupConfig_basic = `
resource "softlayer_lb_local" "local_lb_01" {
    connections = 15000
    datacenter = "sng01"
    ha_enabled = false
}

resource "softlayer_lb_local_service_group" "http_sg" {
    load_balancer_id = "${softlayer_lb_local.local_lb_01.id}"
    allocation = 100
    port = 80
    routing_method = "ROUND_ROBIN"
    routing_type = "HTTP"
}

resource "softlayer_scale_group" "sample-http-cluster" {
    name = "sample-http-cluster"
    regional_group = "as-sgp-central-1"
    cooldown = 30
    minimum_member_count = 1
    maximum_member_count = 10
    termination_policy = "CLOSEST_TO_NEXT_CHARGE"
    virtual_server_id = "${softlayer_lb_local_service_group.http_sg.id}"
    port = 8080
    health_check = {
        type = "HTTP"
    }
    virtual_guest_member_template = {
        name = "test-VM"
        domain = "example.com"
        cpu = 1
        ram = 4096
        public_network_speed = 1000
        hourly_billing = true
        image = "DEBIAN_7_64"
        local_disk = false
        disks = [25,100]
        datacenter = "sng01"
        post_install_script_uri = ""
        user_data = "#!/bin/bash"
    }
}`

const testAccCheckSoftLayerScaleGroupConfig_updated = `
resource "softlayer_lb_local" "local_lb_01" {
    connections = 15000
    datacenter = "sng01"
    ha_enabled = false
}

resource "softlayer_lb_local_service_group" "http_sg" {
    load_balancer_id = "${softlayer_lb_local.local_lb_01.id}"
    allocation = 100
    port = 80
    routing_method = "ROUND_ROBIN"
    routing_type = "HTTP"
}

resource "softlayer_scale_group" "sample-http-cluster" {
    name = "changed_name"
    regional_group = "as-sgp-central-1"
    cooldown = 35
    minimum_member_count = 2
    maximum_member_count = 12
    termination_policy = "NEWEST"
    virtual_server_id = "${softlayer_lb_local_service_group.http_sg.id}"
    port = 9090
    health_check = {
        type = "HTTP-CUSTOM"
        custom_method = "GET"
        custom_request = "/healthcheck"
        custom_response = 200
    }
    virtual_guest_member_template = {
        name = "example-VM"
        domain = "test.com"
        cpu = 2
        ram = 8192
        public_network_speed = 100
        hourly_billing = true
        image = "CENTOS_7_64"
        local_disk = false
        disks = [25,100]
        datacenter = "sng01"
        post_install_script_uri = "http://localhost/index.html"
        user_data = "#!/bin/bash"
    }
}`
