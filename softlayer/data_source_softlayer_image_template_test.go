package softlayer

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccSoftLayerImageTemplateDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSoftLayerImageTemplateDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.softlayer_image_template.tfacc_img_tmpl",
						"name",
						"jumpbox",
					),
					resource.TestCheckResourceAttr(
						"data.softlayer_image_template.tfacc_img_tmpl",
						"global_id",
						"9e039233-9160-4172-8c5a-ed909a008a93",
					),
				),
			},
		},
	})
}

// The datasource to apply
const testAccCheckSoftLayerImageTemplateDataSourceConfig_basic = `

data "softlayer_image_template" "tfacc_img_tmpl" {
    name = "jumpbox"
}
`
