package softlayer

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
	"strconv"
	"testing"
)

func TestAccSoftLayerNetworkStorage_Read(t *testing.T) {
	var netStore datatypes.Network_Storage

	t.Log("Running Resource Test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerNetworkStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: configNetworkStorageBasic,
				Check: resource.ComposeTestCheckFunc(
					testCheckSoftlayerNetworkStorageExists("softlayer_network_storage.test-iscsi", &netStore),
				),
			},
		},
	})
}

func testCheckSoftlayerNetworkStorageExists(name string, netStore *datatypes.Network_Storage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		//Checking the Terraform created state for the resource.
		rsrc, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		//Did Terraform create a resource ID for the store?
		if rsrc.Primary.ID == "" {
			return errors.New("No NetworkStorage ID is set")
		}

		//Convert the id into an Integer.
		id, err := strconv.Atoi(rsrc.Primary.ID)

		if err != nil {
			return err
		}

		//Manually check Softlayer for the NetworkStorage
		service := services.GetNetworkStorageService(testAccProvider.Meta().(*session.Session))
		storage, err := service.Id(id).GetObject()

		if err != nil {
			return err
		}

		if *storage.Id != id {
			return errors.New("NetworkStorage not found")
		}

		*netStore = storage

		return nil

	}
}

func testAccCheckSoftLayerNetworkStorageDestroy(s *terraform.State) error {
	service := services.GetNetworkStorageService(testAccProvider.Meta().(*session.Session))
	for _, rsrc := range s.RootModule().Resources {
		if rsrc.Type != "softlayer_network_storage" {
			continue
		}

		id, _ := strconv.Atoi(rsrc.Primary.ID)

		//Try to find the storage.
		_, err := service.Id(id).GetObject()

		// Wait
		if err != nil {
			if apiErr, ok := err.(sl.Error); !ok || apiErr.StatusCode != 404 {
				return fmt.Errorf(
					"Error waiting for (%d) to be destroyed: %s",
					id, err)
			}
		}
	}
	return nil
}

const configNetworkStorageBasic = `
resource "softlayer_network_storage" "test-iscsi" {
	capacity_gb = 20
	datacenter = "dal09"
	iops = 0.25
	tier = "endurance"
	nas_type = "block"
	notes = "terraform_test"
	os_type = "LINUX"
}
`
