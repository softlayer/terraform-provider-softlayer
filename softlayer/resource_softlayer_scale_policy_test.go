package softlayer

import (
	"fmt"
	"strconv"
	"testing"

	datatypes "github.com/TheWeatherCompany/softlayer-go/data_types"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"strings"
	"time"
)

func TestAccSoftLayerScalePolicy_Basic(t *testing.T) {
	var scalepolicy datatypes.SoftLayer_Scale_Policy
	
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: testAccCheckSoftLayerScalePolicyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config:  testAccCheckSoftLayerScalePolicyConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerScalePolicyExists("softlayer_scale_policy.sample-http-cluster-policy", &scalepolicy),
					testAccCheckSoftLayerScalePolicyAttributes(&scalepolicy),
					resource.TestCheckResourceAttr(
						"softlayer_scale_policy.sample-http-cluster-policy", "name", "sample-http-cluster-policy"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_policy.sample-http-cluster-policy", "scale_type", "RELATIVE"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_policy.sample-http-cluster-policy", "scale_amount", "1"),
					resource.TestCheckResourceAttr(
						"softlayer_scale_policy.sample-http-cluster-policy", "cooldown", "30"),
					resource.TestCheckResourceAttr(
                                                "softlayer_scale_policy.sample-http-cluster-policy", "triggers.#", "3"),
                                        testAccCheckSoftLayerScalePolicyContainsRepeatingTriggers(&scalepolicy, 2, "0 1 ? * MON,WED *"),
                                        testAccCheckSoftLayerScalePolicyContainsResourceUseTriggers(&scalepolicy, 120, "80"),
                                        testAccCheckSoftLayerScalePolicyContainsOneTimeTriggers(&scalepolicy, testOnetimeTriggerDate),
				),
			},
			
			resource.TestStep{
                                Config: testAccCheckSoftLayerScalePolicyConfig_updated,
                                Check: resource.ComposeTestCheckFunc(
                                        testAccCheckSoftLayerScalePolicyExists("softlayer_scale_policy.sample-http-cluster-policy", &scalepolicy),
                                        resource.TestCheckResourceAttr(
                                                "softlayer_scale_policy.sample-http-cluster-policy", "name", "changed-name"),
                                        resource.TestCheckResourceAttr(
                                                "softlayer_scale_policy.sample-http-cluster-policy", "scale_type", "ABSOLUTE"),        
                                        resource.TestCheckResourceAttr(
                                                "softlayer_scale_policy.sample-http-cluster-policy", "scale_amount", "2"),
                                        resource.TestCheckResourceAttr(
                                                "softlayer_scale_policy.sample-http-cluster-policy", "cooldown", "35"),
                                        resource.TestCheckResourceAttr(
                                                "softlayer_scale_policy.sample-http-cluster-policy", "triggers.#", "3"),
                                        testAccCheckSoftLayerScalePolicyContainsRepeatingTriggers(&scalepolicy, 2, "0 1 ? * MON,WED,SAT *"),
                                        testAccCheckSoftLayerScalePolicyContainsResourceUseTriggers(&scalepolicy, 130, "90"),
                                        testAccCheckSoftLayerScalePolicyContainsOneTimeTriggers(&scalepolicy, testOnetimeTriggerUpdatedDate),
                                ),
                        },
		},
	})
}

func testAccCheckSoftLayerScalePolicyDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client).scalePolicyService

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "softlayer_scale_policy" {
			continue
		}

		scalepolicyId, _ := strconv.Atoi(rs.Primary.ID)

		// Try to find the key
		_, err := client.GetObject(scalepolicyId)

		if err != nil && !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("Error waiting for Auto Scale Policy (%s) to be destroyed: %s", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckSoftLayerScalePolicyContainsResourceUseTriggers(scalePolicy *datatypes.SoftLayer_Scale_Policy, period int, value string) resource.TestCheckFunc {
        return func(s *terraform.State) error {
                found := false

                for _, scaleResourceUseTrigger := range scalePolicy.ResourceUseTriggers {
                        for _, scaleResourceUseWatch := range scaleResourceUseTrigger.Watches {
                                if scaleResourceUseWatch.Metric == "host.cpu.percent" && scaleResourceUseWatch.Operator == ">" &&
                                        scaleResourceUseWatch.Period == period && scaleResourceUseWatch.Value == value {
                                        found = true
                                        break
                                }
                        }
                }

                if !found {
                        return fmt.Errorf("Resource use trigger not found in scale policy")

                }

                return nil
        }
}

func testAccCheckSoftLayerScalePolicyContainsRepeatingTriggers(scalePolicy *datatypes.SoftLayer_Scale_Policy, typeId int, schedule string) resource.TestCheckFunc {
        return func(s *terraform.State) error {
                found := false

                for _, scaleRepeatingTrigger := range scalePolicy.RepeatingTriggers {
                        if scaleRepeatingTrigger.TypeId == typeId && scaleRepeatingTrigger.Schedule == schedule {
                                found = true
                                break
                        }
                }

                if !found {
                        return fmt.Errorf("Repeating trigger %d with schedule %s not found in scale policy", typeId, schedule)

                }

                return nil
        }
}

func testAccCheckSoftLayerScalePolicyContainsOneTimeTriggers(scalePolicy *datatypes.SoftLayer_Scale_Policy, testOnetimeTriggerDate string) resource.TestCheckFunc {
        return func(s *terraform.State) error {
                found := false
                const SoftLayerTimeFormat = "2006-01-02T15:04:05-07:00"
                estLoc, _ := time.LoadLocation("EST")

                for _, scaleOneTimeTrigger := range scalePolicy.OneTimeTriggers {
                        if scaleOneTimeTrigger.Date.In(estLoc).Format(SoftLayerTimeFormat) == testOnetimeTriggerDate {
                                found = true
                                break
                        }
                }

                if !found {
                        return fmt.Errorf("One time trigger with date %s not found in scale policy", testOnetimeTriggerDate)
                }

                return nil

        }
}

func testAccCheckSoftLayerScalePolicyAttributes(scalepolicy *datatypes.SoftLayer_Scale_Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if scalepolicy.Name != "sample-http-cluster-policy" {
			return fmt.Errorf("Bad name: %s", scalepolicy.Name)
		}

		return nil
	}
}

func testAccCheckSoftLayerScalePolicyExists(n string, scalepolicy *datatypes.SoftLayer_Scale_Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		scalepolicyId, _ := strconv.Atoi(rs.Primary.ID)

		client := testAccProvider.Meta().(*Client).scalePolicyService
		foundScalePolicy, err := client.GetObject(scalepolicyId)

		if err != nil {
			return err
		}

		if strconv.Itoa(int(foundScalePolicy.Id)) != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		*scalepolicy = foundScalePolicy

		return nil
	}
}

var testAccCheckSoftLayerScalePolicyConfig_basic = fmt.Sprintf( `
resource "softlayer_scale_group" "sample-http-cluster" {
    name = "sample-http-cluster"
    regional_group = "as-sgp-central-1" 
    cooldown = 30
    minimum_member_count = 1
    maximum_member_count = 10
    termination_policy = "CLOSEST_TO_NEXT_CHARGE"
    virtual_server_id = 267513
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
        region = "sng01"
    }
}

resource "softlayer_scale_policy" "sample-http-cluster-policy" {
    name = "sample-http-cluster-policy"
    scale_type = "RELATIVE"
    scale_amount = 1
    cooldown = 30
    scale_group_id = "${softlayer_scale_group.sample-http-cluster.id}"
    triggers = {
        type = "RESOURCE_USE"
        watches = {

                    metric = "host.cpu.percent"
                    operator = ">"
                    value = "80"
                    period = 120
        }
    }
    triggers = {
        type = "ONE_TIME"
        date = "%s"
    }
    triggers = {
        type = "REPEATING"
        schedule = "0 1 ? * MON,WED *"
    }
    
}`, testOnetimeTriggerDate)

const SoftLayerTimeFormat = string("2006-01-02T15:04:05-07:00")

var estLoc, _ = time.LoadLocation("EST")

var testOnetimeTriggerDate = time.Now().In(estLoc).AddDate(0, 0, 1).Format(SoftLayerTimeFormat)

var testAccCheckSoftLayerScalePolicyConfig_updated = fmt.Sprintf(`
resource "softlayer_scale_group" "sample-http-cluster" {
    name = "sample-http-cluster"
    regional_group = "as-sgp-central-1"
    cooldown = 30
    minimum_member_count = 1
    maximum_member_count = 10
    termination_policy = "CLOSEST_TO_NEXT_CHARGE"
    virtual_server_id = 267513
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
        region = "sng01"
    }
}
resource "softlayer_scale_policy" "sample-http-cluster-policy" {
    name = "changed-name"
    scale_type = "ABSOLUTE"
    scale_amount = 2
    cooldown = 35
    scale_group_id = "${softlayer_scale_group.sample-http-cluster.id}"
    triggers = {
        type = "RESOURCE_USE"
        watches = {

                    metric = "host.cpu.percent"
                    operator = ">"
                    value = "90"
                    period = 130
        }
    }
    triggers = {
        type = "REPEATING"
        schedule = "0 1 ? * MON,WED,SAT *"
    }
    triggers = {
        type = "ONE_TIME"
        date = "%s"
    }
}`, testOnetimeTriggerUpdatedDate)

 var testOnetimeTriggerUpdatedDate = time.Now().In(estLoc).AddDate(0, 0, 2).Format(SoftLayerTimeFormat)
