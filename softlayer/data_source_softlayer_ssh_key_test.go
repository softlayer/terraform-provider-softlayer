package softlayer

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccSoftLayerSSHKeyDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerSSHKeyDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.softlayer_ssh_key.tfacc_ssh_key", "public_key", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDRSafsDMQWj12uQm+tgdjnLcuojMpmAfqiEltz7eZlRn77ivYgWGFyWq3WW+DgeW5QG/W0Cq0C6cB8HhP+hGpF/YTzEplPe5GMksH15fN15KVDUEo1F6UBcz80AAMjxNK3Qx8zItR3xbUW7BuNz7IVp6x4wr9DHNO+Yqx0MJpPI6C8wLpdL8ihNW4SSWMiKrzkdENd2wHqn+YoCgJrT2rFM1lvapCfVQrbhiJR2Rk99Fz8sDCpUZTvCmS21wihwo973GpllkjaN5sBAZkimetG+Te/iPRJSGXy/sF2krSf1REs9AdleA5o+C7iVFll0Q1rjMdKrGTcOyjj8BY6xuuj"),
					resource.TestCheckResourceAttr("data.softlayer_ssh_key.tfacc_ssh_key", "notes", "Public ssh key for terraform acceptance test"),
					resource.TestMatchResourceAttr("data.softlayer_ssh_key.tfacc_ssh_key", "fingerprint", regexp.MustCompile("^[0-9a-f]{2}:")),
				),
			},
		},
	})
}

// The datasource to apply
const testAccCheckSoftLayerSSHKeyDataSourceConfig_basic = `

data "softlayer_ssh_key" "tfacc_ssh_key" {
    label = "tfacc ssh key"
}
`

/*
 * Note: Before running this test, apply the following configuration into the test SoftLayer account:

	resource "softlayer_ssh_key" "ssh_key_1" {
		name = "tfacc ssh key"
		notes = "Public ssh key for terraform acceptance test"
		public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDRSafsDMQWj12uQm+tgdjnLcuojMpmAfqiEltz7eZlRn77ivYgWGFyWq3WW+DgeW5QG/W0Cq0C6cB8HhP+hGpF/YTzEplPe5GMksH15fN15KVDUEo1F6UBcz80AAMjxNK3Qx8zItR3xbUW7BuNz7IVp6x4wr9DHNO+Yqx0MJpPI6C8wLpdL8ihNW4SSWMiKrzkdENd2wHqn+YoCgJrT2rFM1lvapCfVQrbhiJR2Rk99Fz8sDCpUZTvCmS21wihwo973GpllkjaN5sBAZkimetG+Te/iPRJSGXy/sF2krSf1REs9AdleA5o+C7iVFll0Q1rjMdKrGTcOyjj8BY6xuuj"
	}

*/
