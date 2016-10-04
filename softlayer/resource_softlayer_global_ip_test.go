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
)

func TestAccSoftLayerGlobalIp_Basic(t *testing.T) {
        var globalIp datatypes.Network_Subnet_IpAddress_Global

        resource.Test(t, resource.TestCase{
                PreCheck:     func() { testAccPreCheck(t) },
                Providers:    testAccProviders,
                CheckDestroy: testAccCheckSoftLayerGlobalIpDestroy,
                Steps: []resource.TestStep{
                        resource.TestStep{
                                Config: testAccCheckSoftLayerGlobalIpConfig_basic,
                                Check: resource.ComposeTestCheckFunc(
                                        testAccCheckSoftLayerGlobalIpExists("softlayer_global_ip.test-global-ip", &globalIp),
                                        testAccCheckSoftLayerGlobalIpAttributes(&globalIp),
                                        resource.TestCheckResourceAttr(
                                                "softlayer_global_ip.test-global-ip", "routes_to", "119.81.82.163"),
                                ),
                        },                        
                },
        })
}

func testAccCheckSoftLayerGlobalIpDestroy(s *terraform.State) error {
        service := services.GetNetworkSubnetIpAddressGlobalService(testAccProvider.Meta().(*session.Session))

        for _, rs := range s.RootModule().Resources {
                if rs.Type != "softlayer_global_ip" {
                        continue
                }

                globalIpId, _ := strconv.Atoi(rs.Primary.ID)

                // Try to find the global ip
                _, err := service.Id(globalIpId).GetObject()

                if err == nil {
                        return fmt.Errorf("Global Ip still exists")
                }
        }

        return nil
}

func testAccCheckSoftLayerGlobalIpAttributes(globalIp *datatypes.Network_Subnet_IpAddress_Global) resource.TestCheckFunc {
        return func(s *terraform.State) error {

                if *globalIp.DestinationIpAddress.IpAddress != "119.81.82.160" {
                        return fmt.Errorf("Bad destination ip address: %s", *globalIp.DestinationIpAddress.IpAddress)
                }

                return nil
        }
}

func testAccCheckSoftLayerGlobalIpExists(n string, globalIp *datatypes.Network_Subnet_IpAddress_Global) resource.TestCheckFunc {
        return func(s *terraform.State) error {
                rs, ok := s.RootModule().Resources[n]

                if !ok {
                        return fmt.Errorf("Not found: %s", n)
                }

                if rs.Primary.ID == "" {
                        return fmt.Errorf("No Record ID is set")
                }

                globalIpId, _ := strconv.Atoi(rs.Primary.ID)

                service := services.GetNetworkSubnetIpAddressGlobalService(testAccProvider.Meta().(*session.Session))
                foundGlobalIp, err := service.Id(globalIpId).GetObject()

                if err != nil {
                        return err
                }

                if strconv.Itoa(int(*foundGlobalIp.Id)) != rs.Primary.ID {
                        return fmt.Errorf("Record not found")
                }

                *globalIp = foundGlobalIp

                return nil
        }
}

const testAccCheckSoftLayerGlobalIpConfig_basic = `
resource "softlayer_global_ip" "test-global-ip" {
    routes_to = "119.81.82.160"
}`
