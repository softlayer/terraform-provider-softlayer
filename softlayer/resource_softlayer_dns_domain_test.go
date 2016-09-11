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
	"github.com/softlayer/softlayer-go/sl"
)

func TestAccSoftLayerDnsDomain_Basic(t *testing.T) {
	var dns_domain datatypes.Dns_Domain

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerDnsDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(config, domainName1, target1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerDnsDomainExists("softlayer_dns_domain.acceptance_test_dns_domain-1", &dns_domain),
					testAccCheckSoftLayerDnsDomainAttributes(&dns_domain),
					saveSoftLayerDnsDomainId(&dns_domain, &firstDnsId),
					resource.TestCheckResourceAttr(
						"softlayer_dns_domain.acceptance_test_dns_domain-1", "name", domainName1),
					resource.TestCheckResourceAttr(
						"softlayer_dns_domain.acceptance_test_dns_domain-1", "target", target1),
				),
				Destroy: false,
			},
			{
				Config: fmt.Sprintf(config, domainName2, target1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerDnsDomainExists("softlayer_dns_domain.acceptance_test_dns_domain-1", &dns_domain),
					testAccCheckSoftLayerDnsDomainAttributes(&dns_domain),
					testAccCheckSoftLayerDnsDomainRecordDomainId(
						"softlayer_dns_domain.acceptance_test_dns_domain-1", &dns_domain),
					resource.TestCheckResourceAttr(
						"softlayer_dns_domain.acceptance_test_dns_domain-1", "name", domainName2),
					resource.TestCheckResourceAttr(
						"softlayer_dns_domain.acceptance_test_dns_domain-1", "target", target1),
					testAccCheckSoftLayerDnsDomainChanged(&dns_domain),
				),
				Destroy: false,
			},
			{
				Config: fmt.Sprintf(config, domainName2, target2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerDnsDomainExists("softlayer_dns_domain.acceptance_test_dns_domain-1", &dns_domain),
					testAccCheckSoftLayerDnsDomainAttributes(&dns_domain),
					testAccCheckSoftLayerDnsDomainRecordDomainId(
						"softlayer_dns_domain.acceptance_test_dns_domain-1", &dns_domain),
					resource.TestCheckResourceAttr(
						"softlayer_dns_domain.acceptance_test_dns_domain-1", "name", domainName2),
					resource.TestCheckResourceAttr(
						"softlayer_dns_domain.acceptance_test_dns_domain-1", "target", target2),
				),
				Destroy: false,
			},
		},
	})
}

func testAccCheckSoftLayerDnsDomainDestroy(s *terraform.State) error {
	service := services.GetDnsDomainService(testAccProvider.Meta().(*session.Session))

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "softlayer_dns_domain" {
			continue
		}

		dnsId, _ := strconv.Atoi(rs.Primary.ID)

		// Try to find the domain
		_, err := service.Id(dnsId).GetObject()

		if err == nil {
			return fmt.Errorf("Dns Domain with id %d still exists", dnsId)
		}
	}

	return nil
}

func testAccCheckSoftLayerDnsDomainRecordDomainId(n string, dns_domain *datatypes.Dns_Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		attrs := rs.Primary.Attributes
		recordsTotal, _ := strconv.Atoi(attrs["records.#"])
		for i := 0; i < recordsTotal; i++ {
			recordDomainId, _ := strconv.Atoi(attrs[fmt.Sprintf("records.%d.domain_id", i)])
			if *dns_domain.Id != recordDomainId {
				return fmt.Errorf(
					"Dns domain id (%d) and Dns domain record domain id (%d) should be equal",
					*dns_domain.Id, recordDomainId,
				)
			}
		}

		return nil
	}
}

func testAccCheckSoftLayerDnsDomainAttributes(dns *datatypes.Dns_Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if name := sl.Get(dns.Name); name == "" {
			return errors.New("Empty dns domain name")
		}

		// find a record with host @; that will have the current target.
		foundTarget := false
		for _, record := range dns.ResourceRecords {
			if *record.Type == "a" && *record.Host == "@" {
				foundTarget = true
				break
			}
		}

		if !foundTarget {
			return fmt.Errorf("Target record not found for dns domain %s (%d)", sl.Get(dns.Name), sl.Get(dns.Id))
		}

		if id := sl.Get(dns.Id); id == 0 {
			return fmt.Errorf("Bad dns domain id: %d", id)
		}

		return nil
	}
}

func saveSoftLayerDnsDomainId(dns *datatypes.Dns_Domain, id_holder *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		*id_holder = *dns.Id

		return nil
	}
}

func testAccCheckSoftLayerDnsDomainChanged(dns *datatypes.Dns_Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		service := services.GetDnsDomainService(testAccProvider.Meta().(*session.Session))

		_, err := service.Id(firstDnsId).Mask(
			"id,name,updateDate,resourceRecords",
		).GetObject()
		if err == nil {
			return fmt.Errorf("Dns domain with id %d still exists", firstDnsId)
		}

		return nil
	}
}

func testAccCheckSoftLayerDnsDomainExists(n string, dns_domain *datatypes.Dns_Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Record ID is set")
		}

		dns_id, _ := strconv.Atoi(rs.Primary.ID)

		service := services.GetDnsDomainService(testAccProvider.Meta().(*session.Session))
		found_domain, err := service.Id(dns_id).Mask(
			"id,name,updateDate,resourceRecords",
		).GetObject()

		if err != nil {
			return err
		}

		if strconv.Itoa(int(*found_domain.Id)) != rs.Primary.ID {
			return errors.New("Record not found")
		}

		*dns_domain = found_domain

		return nil
	}
}

var config = `
resource "softlayer_dns_domain" "acceptance_test_dns_domain-1" {
	name = "%s"
	target = "%s"
}
`

var domainName1 = "zxczcxzxc.com"
var domainName2 = "vbnvnvbnv.com"
var target1 = "172.16.0.100"
var target2 = "172.16.0.101"
var firstDnsId = 0
