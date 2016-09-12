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

func TestAccSoftLayerDnsDomainRecord_Basic(t *testing.T) {
	var dns_domain datatypes.Dns_Domain
	var dns_domain_record datatypes.Dns_Domain_ResourceRecord

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerDnsDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSoftLayerDnsDomainRecordConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerDnsDomainExists("softlayer_dns_domain.test_dns_domain_records", &dns_domain),
					testAccCheckSoftLayerDnsDomainRecordExists("softlayer_dns_domain_record.recordA", &dns_domain_record),
					resource.TestCheckResourceAttr("softlayer_dns_domain_record.recordA", "data", "127.0.0.1"),
					resource.TestCheckResourceAttr("softlayer_dns_domain_record.recordA", "expire", "900"),
					resource.TestCheckResourceAttr("softlayer_dns_domain_record.recordA", "minimum_ttl", "90"),
					resource.TestCheckResourceAttr("softlayer_dns_domain_record.recordA", "mx_priority", "1"),
					resource.TestCheckResourceAttr("softlayer_dns_domain_record.recordA", "refresh", "1"),
					resource.TestCheckResourceAttr("softlayer_dns_domain_record.recordA", "host", "hosta.com"),
					resource.TestCheckResourceAttr("softlayer_dns_domain_record.recordA", "responsible_person", "user@softlayer.com"),
					resource.TestCheckResourceAttr("softlayer_dns_domain_record.recordA", "ttl", "900"),
					resource.TestCheckResourceAttr("softlayer_dns_domain_record.recordA", "retry", "1"),
					resource.TestCheckResourceAttr("softlayer_dns_domain_record.recordA", "type", "a"),
				),
			},
		},
	})
}

func TestAccSoftLayerDnsDomainRecord_Types(t *testing.T) {
	var dns_domain datatypes.Dns_Domain
	var dns_domain_record datatypes.Dns_Domain_ResourceRecord

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerDnsDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckSoftLayerDnsDomainRecordConfig_all_types, "_tcp"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerDnsDomainExists("softlayer_dns_domain.test_dns_domain_record_types", &dns_domain),
					testAccCheckSoftLayerDnsDomainRecordExists("softlayer_dns_domain_record.recordA", &dns_domain_record),
					testAccCheckSoftLayerDnsDomainRecordExists("softlayer_dns_domain_record.recordAAAA", &dns_domain_record),
					testAccCheckSoftLayerDnsDomainRecordExists("softlayer_dns_domain_record.recordCNAME", &dns_domain_record),
					testAccCheckSoftLayerDnsDomainRecordExists("softlayer_dns_domain_record.recordMX", &dns_domain_record),
					testAccCheckSoftLayerDnsDomainRecordExists("softlayer_dns_domain_record.recordSPF", &dns_domain_record),
					testAccCheckSoftLayerDnsDomainRecordExists("softlayer_dns_domain_record.recordTXT", &dns_domain_record),
					testAccCheckSoftLayerDnsDomainRecordExists("softlayer_dns_domain_record.recordSRV", &dns_domain_record),
				),
				Destroy: false,
			},
			{
				Config: fmt.Sprintf(testAccCheckSoftLayerDnsDomainRecordConfig_all_types, "_udp"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSoftLayerDnsDomainExists("softlayer_dns_domain.test_dns_domain_record_types", &dns_domain),
					resource.TestCheckResourceAttr("softlayer_dns_domain_record.recordSRV", "protocol", "_udp"),
				),
				Destroy: false,
			},
		},
	})
}

func testAccCheckSoftLayerDnsDomainRecordExists(n string, dns_domain_record *datatypes.Dns_Domain_ResourceRecord) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Record ID is set")
		}

		dns_id, _ := strconv.Atoi(rs.Primary.ID)

		service := services.GetDnsDomainResourceRecordService(testAccProvider.Meta().(*session.Session))
		found_domain_record, err := service.Id(dns_id).GetObject()

		if err != nil {
			return err
		}

		if strconv.Itoa(int(*found_domain_record.Id)) != rs.Primary.ID {
			return fmt.Errorf("Record %d not found", dns_id)
		}

		*dns_domain_record = found_domain_record

		return nil
	}
}

var testAccCheckSoftLayerDnsDomainRecordConfig_basic = `
resource "softlayer_dns_domain" "test_dns_domain_records" {
	name = "domain.records.com"
	target = "172.16.0.100"
}

resource "softlayer_dns_domain_record" "recordA" {
    data = "127.0.0.1"
    domain_id = "${softlayer_dns_domain.test_dns_domain_records.id}"
    expire = 900
    minimum_ttl = 90
    mx_priority = 1
    refresh = 1
    host = "hosta.com"
    responsible_person = "user@softlayer.com"
    ttl = 900
    retry = 1
    type = "a"
}
`
var testAccCheckSoftLayerDnsDomainRecordConfig_all_types = `
resource "softlayer_dns_domain" "test_dns_domain_record_types" {
	name = "domain.record.types.com"
	target = "172.16.0.100"
}

resource "softlayer_dns_domain_record" "recordA" {
    data = "127.0.0.1"
    domain_id = "${softlayer_dns_domain.test_dns_domain_record_types.id}"
    host = "hosta.com"
    responsible_person = "user@softlayer.com"
    ttl = 900
    type = "a"
}

resource "softlayer_dns_domain_record" "recordAAAA" {
    data = "fe80:0000:0000:0000:0202:b3ff:fe1e:8329"
    domain_id = "${softlayer_dns_domain.test_dns_domain_record_types.id}"
    host = "hosta-2.com"
    responsible_person = "user2changed@softlayer.com"
    ttl = 1000
    type = "aaaa"
}

resource "softlayer_dns_domain_record" "recordCNAME" {
    data = "testsssaaaass.com"
    domain_id = "${softlayer_dns_domain.test_dns_domain_record_types.id}"
    host = "hosta-cname.com"
    responsible_person = "user@softlayer.com"
    ttl = 900
    type = "cname"
}

resource "softlayer_dns_domain_record" "recordMX" {
    data = "email.example.com"
    domain_id = "${softlayer_dns_domain.test_dns_domain_record_types.id}"
    host = "hosta-mx.com"
    responsible_person = "user@softlayer.com"
    ttl = 900
    type = "mx"
}

resource "softlayer_dns_domain_record" "recordSPF" {
    data = "v=spf1 mx:mail.example.org ~all"
    domain_id = "${softlayer_dns_domain.test_dns_domain_record_types.id}"
    host = "hosta-spf"
    responsible_person = "user@softlayer.com"
    ttl = 900
    type = "spf"
}

resource "softlayer_dns_domain_record" "recordTXT" {
    data = "127.0.0.1"
    domain_id = "${softlayer_dns_domain.test_dns_domain_record_types.id}"
    host = "hosta-txt.com"
    responsible_person = "user@softlayer.com"
    ttl = 900
    type = "txt"
}

resource "softlayer_dns_domain_record" "recordSRV" {
    data = "ns1.example.org"
    domain_id = "${softlayer_dns_domain.test_dns_domain_record_types.id}"
    host = "hosta-srv.com"
    responsible_person = "user@softlayer.com"
    ttl = 900
    type = "srv"
	port = 8080
	priority = 3
	protocol = "%s"
	weight = 3
	service = "_mail"
}
`
