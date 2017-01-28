package softlayer

import (
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

func TestAccSoftLayerBlockStorage_Endurance_Basic(t *testing.T) {
	var netStore datatypes.Network_Storage

	t.Log("Running Basic Endurance Storage Test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerBlockStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: configEnduranceBlockStorageBasic,
				Check: resource.ComposeTestCheckFunc(
					testCheckSoftlayerBlockStorageExists("softlayer_block_storage.test-basic-endurance", &netStore),
				),
			},
		},
	})
}

func testCheckSoftlayerBlockStorageExists(name string, netStore *datatypes.Network_Storage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		//Checking the Terraform created state for the resource.
		rsrc, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		//Did Terraform create a resource ID for the store?
		if rsrc.Primary.ID == "" {
			return fmt.Errorf("No BlockStorage ID is set")
		}

		//Convert the id into an Integer.
		id, err := strconv.Atoi(rsrc.Primary.ID)

		if err != nil {
			return err
		}

		//Manually check Softlayer for the BlockStorage
		service := services.GetNetworkStorageService(testAccProvider.Meta().(*session.Session))
		storage, err := service.Id(id).GetObject()

		if err != nil {
			return err
		}

		if *storage.Id != id {
			return fmt.Errorf("BlockStorage not found")
		}

		*netStore = storage

		return nil

	}
}

func testAccCheckSoftLayerBlockStorageDestroy(s *terraform.State) error {
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

const configEnduranceBlockStorageBasic = `
resource "softlayer_block_storage" "test-basic-endurance" {
	capacity_gb = 20
	datacenter = "dal09"
	iops = 0.25
	tier = "endurance"
	os_type = "LINUX"
}
`
