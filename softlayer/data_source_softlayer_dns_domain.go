package softlayer

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
)

func dataSourceSoftLayerDnsDomain() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSoftLayerDnsDomainRead,

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Description: "A domain record's internal identifier",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"name": &schema.Schema{
				Description: "The name of the domain",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceSoftLayerDnsDomainRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetAccountService(sess)

	name := d.Get("name").(string)

	names, err := service.
		Filter(filter.Build(filter.Path("domains.name").Eq(name))).
		Mask("id,name").
		GetDomains()

	if err != nil {
		return fmt.Errorf("Error retrieving domain: %s", err)
	}

	if len(names) == 0 {
		return fmt.Errorf("No domain found with name [%s]", name)
	}

	d.SetId(fmt.Sprintf("%d", *names[0].Id))
	return nil
}
