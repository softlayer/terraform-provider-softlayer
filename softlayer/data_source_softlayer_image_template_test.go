package softlayer

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccSoftLayerImageTemplateDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Tests looking up private or shared images
			{
				Config: testAccCheckSoftLayerImageTemplateDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.softlayer_image_template.tfacc_img_tmpl",
						"name",
						"jumpbox",
					),
					resource.TestMatchResourceAttr(
						"data.softlayer_image_template.tfacc_img_tmpl",
						"id",
						regexp.MustCompile("^[0-9]+$"),
					),
				),
			},
			// Tests looking up a public image
			{
				Config: testAccCheckSoftLayerImageTemplateDataSourceConfig_basic2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.softlayer_image_template.tfacc_img_tmpl",
						"name",
						"RightImage_Ubuntu_12.04_amd64_v13.5",
					),
					resource.TestMatchResourceAttr(
						"data.softlayer_image_template.tfacc_img_tmpl",
						"id",
						regexp.MustCompile("^[0-9]+$"),
					),
				),
			},
		},
	})
}

const testAccCheckSoftLayerImageTemplateDataSourceConfig_basic = `
data "softlayer_image_template" "tfacc_img_tmpl" {
    name = "jumpbox"
}
`

const testAccCheckSoftLayerImageTemplateDataSourceConfig_basic2 = `
data "softlayer_image_template" "tfacc_img_tmpl" {
    name = "RightImage_Ubuntu_12.04_amd64_v13.5"
}
`
