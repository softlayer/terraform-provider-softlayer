package softlayer

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
)

func dataSourceSoftLayerImageTemplate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSoftLayerImageTemplateRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The internal id of the image template",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"name": {
				Description: "The name of this image template",
				Type:        schema.TypeString,
				Required:    true,
			},

			"global_id": {
				Description: "The global identifier of the image template",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceSoftLayerImageTemplateRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetAccountService(sess)

	name := d.Get("name").(string)

	imageTemplates, err := service.
		Filter(filter.Build(filter.Path("name").Eq(name))).
		Mask("id,name,globalIdentifier").
		GetBlockDeviceTemplateGroups()

	if err != nil {
		return fmt.Errorf("Error looking up image template [%s]: %s", name, err)
	}

	if len(imageTemplates) == 0 {
		return fmt.Errorf("No image template found with name [%s]", name)
	}

	imageTemplate := imageTemplates[0]

	d.SetId(fmt.Sprintf("%d", *imageTemplate.Id))
	d.Set("name", name)
	d.Set("global_id", *imageTemplate.GlobalIdentifier)

	return nil
}
