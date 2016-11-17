package softlayer

import (
	"fmt"
	"strconv"
	"testing"

	"crypto/sha1"
	"encoding/hex"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"regexp"
)

func TestAccSoftLayerUser_Basic(t *testing.T) {
	var user datatypes.User_Customer

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerUserDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerUserConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerUserExists("softlayer_user.testuser", &user),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "username", testAccRandomUserName),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "first_name", "first_name"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "last_name", "last_name"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "email", "first.last@example.com"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "company_name", "company_name"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "address1", "1 Main St."),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "address2", "Suite 345"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "city", "Atlanta"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "state", "GA"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "country", "US"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "timezone", "EST"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "user_status", "ACTIVE"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "password", hash(testAccUserPassword)),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "permissions.#", "2"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "has_api_key", "true"),
					resource.TestMatchResourceAttr(
						"softlayer_user.testuser", "api_key", apiKeyRegexp),
				),
			},

			resource.TestStep{
				Config: testAccCheckSoftLayerUserConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "username", testAccRandomUserName),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "first_name", "new_first_name"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "last_name", "new_last_name"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "email", "new_first.new_last@example.com"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "company_name", "new_company_name"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "address1", "1 1st Avenue"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "address2", "Apartment 2"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "city", "Montreal"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "state", "QC"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "country", "CA"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "timezone", "MST"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "user_status", "INACTIVE"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "password", hash(testAccUserPassword)),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "permissions.#", "3"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "has_api_key", "false"),
					resource.TestCheckResourceAttr(
						"softlayer_user.testuser", "api_key", ""),
				),
			},
		},
	})
}

func testAccCheckSoftLayerUserDestroy(s *terraform.State) error {
	client := services.GetUserCustomerService(testAccProvider.Meta().(ProviderConfig).SoftLayerSession())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "softlayer_user" {
			continue
		}

		userID, _ := strconv.Atoi(rs.Primary.ID)

		// Try to find the user
		user, err := client.Id(userID).Mask("userStatusId").GetObject()

		// Users are not immediately deleted, but rather placed into a 'cancel_pending' (1021) status
		if err != nil || *user.UserStatusId != userCustomerCancelStatus {
			return fmt.Errorf("SoftLayer User still exists")
		}
	}

	return nil
}

func testAccCheckSoftLayerUserExists(n string, user *datatypes.User_Customer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		userID, _ := strconv.Atoi(rs.Primary.ID)

		client := services.GetUserCustomerService(testAccProvider.Meta().(ProviderConfig).SoftLayerSession())
		foundUser, err := client.Id(userID).GetObject()

		if err != nil {
			return err
		}

		if strconv.Itoa(int(*foundUser.Id)) != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		*user = foundUser

		return nil
	}
}

var testAccCheckSoftLayerUserConfig_basic = fmt.Sprintf(`
resource "softlayer_user" "testuser" {
    username = "%s"
    first_name = "first_name"
    last_name = "last_name"
    email = "first.last@example.com"
    company_name = "company_name"
    address1 = "1 Main St."
    address2 = "Suite 345"
    city = "Atlanta"
    state = "GA"
    country = "US"
    timezone = "EST"
    password = "%s"
    permissions = [
        "SERVER_ADD",
        "ACCESS_ALL_GUEST"
    ]
    has_api_key = true
}`, testAccRandomUserName, testAccUserPassword)

var testAccCheckSoftLayerUserConfig_updated = fmt.Sprintf(`
resource "softlayer_user" "testuser" {
    username = "%s"
    first_name = "new_first_name"
    last_name = "new_last_name"
    email = "new_first.new_last@example.com"
    company_name = "new_company_name"
    address1 = "1 1st Avenue"
    address2 = "Apartment 2"
    city = "Montreal"
    state = "QC"
    country = "CA"
    timezone = "MST"
    user_status = "INACTIVE"
    password = "%s"
    permissions = [
        "SERVER_ADD",
        "ACCESS_ALL_HARDWARE",
        "TICKET_EDIT"
    ]
    has_api_key = false
}`, testAccRandomUserName, testAccUserPassword)

var testAccRandomUserName = resource.UniqueId()
var testAccUserPassword = "T3stp@ss"
var apiKeyRegexp, _ = regexp.Compile(`\w+`)

// Function used by provider for hashing passwords
func hash(v interface{}) string {
	hash := sha1.Sum([]byte(v.(string)))
	return hex.EncodeToString(hash[:])
}
