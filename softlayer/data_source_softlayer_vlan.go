package softlayer

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
)

func dataSourceSoftLayerVlan() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSoftLayerVlanRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"number", "router_hostname"},
			},

			"number": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"name"},
			},

			"router_hostname": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name"},
			},
		},
	}
}

func dataSourceSoftLayerVlanRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetAccountService(sess)

	name := d.Get("name").(string)
	number := d.Get("number").(int)
	routerHostname := d.Get("router_hostname").(string)

	if name != "" {
		networkVlans, err := service.
			Mask("id").
			Filter(filter.Path("networkVlans.name").Eq(name).Build()).
			GetNetworkVlans()
		if err != nil || len(networkVlans) == 0 {
			return fmt.Errorf("Error obtaining VLAN id: %s", err)
		}

		d.SetId(fmt.Sprintf("%d", *networkVlans[0].Id))
	} else if number != 0 && routerHostname != "" {
		id, err := getVlanId(number, routerHostname, meta)
		if err != nil {
			return fmt.Errorf("Error obtaining VLAN id: %s", err)
		}

		d.SetId(fmt.Sprintf("%d", id))
	} else {
		return errors.New("Missing required properties. Need a VLAN name, or the VLAN's number and router hostname.")
	}

	return nil
}
