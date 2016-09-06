package softlayer

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccSoftLayerObjectStorageAccount_Basic(t *testing.T) {
	var accountName string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerObjectStorageAccountDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerObjectStorageAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerObjectStorageAccountExists("softlayer_objectstorage_account.testacc_foobar", &accountName),
					testAccCheckSoftLayerObjectStorageAccountAttributes(&accountName),
				),
			},
		},
	})
}

func testAccCheckSoftLayerObjectStorageAccountDestroy(s *terraform.State) error {
	return nil
}

func testAccCheckSoftLayerObjectStorageAccountExists(n string, accountName *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		*accountName = rs.Primary.ID

		return nil
	}
}

func testAccCheckSoftLayerObjectStorageAccountAttributes(accountName *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *accountName == "" {
			return fmt.Errorf("No object storage account name")
		}

		return nil
	}
}

var testAccCheckSoftLayerObjectStorageAccountConfig_basic = `
resource "softlayer_objectstorage_account" "testacc_foobar" {
}`
