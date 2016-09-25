package softlayer

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
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
		},
	}
}

func dataSourceSoftLayerImageTemplateRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetAccountService(sess)

	name := d.Get("name").(string)

	imageTemplates, err := service.
		Mask("id,name").
		GetBlockDeviceTemplateGroups()

	if err != nil {
		return fmt.Errorf("Error looking up image template [%s]: %s", name, err)
	}

	if len(imageTemplates) == 0 {
		return errors.New("The SoftLayer account has no image templates.")
	}

	for _, imageTemplate := range imageTemplates {
		if imageTemplate.Name != nil && *imageTemplate.Name == name {
			d.SetId(fmt.Sprintf("%d", *imageTemplate.Id))
			return nil
		}
	}

	return fmt.Errorf("Could not find image template with name [%s]", name)
}
