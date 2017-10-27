package softlayer

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
)

func dataSourceSoftLayerImageTemplate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSoftLayerImageTemplateRead,

		// TODO: based on need add properties for visibility, type of image,
		// notes, size, shared on accounts if needed
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of this image template",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceSoftLayerImageTemplateRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetAccountService(sess)

	name := d.Get("name").(string)

	imageTemplates, err := service.
		Mask("id,name").
		GetBlockDeviceTemplateGroups()
	if err != nil {
		return fmt.Errorf("Error looking up image template [%s]: %s", name, err)
	}

	for _, imageTemplate := range imageTemplates {
		if imageTemplate.Name != nil && *imageTemplate.Name == name {
			d.SetId(fmt.Sprintf("%d", *imageTemplate.Id))
			return nil
		}
	}

	// Image not found among private nor shared images in the account.
	// Looking up in the public images
	templateService := services.GetVirtualGuestBlockDeviceTemplateGroupService(sess)
	pubImageTemplates, err := templateService.
		Mask("id,name").
		Filter(filter.Path("name").Eq(name).Build()).
		GetPublicImages()
	if err != nil {
		return fmt.Errorf("Error looking up image template [%s] among the public images: %s", name, err)
	}

	if len(pubImageTemplates) > 0 {
		imageTemplate := pubImageTemplates[0]
		d.SetId(fmt.Sprintf("%d", *imageTemplate.Id))
		return nil
	}

	return fmt.Errorf("Could not find image template with name [%s]", name)
}
