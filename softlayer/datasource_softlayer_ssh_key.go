package softlayer

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
)

func dataSourceSoftLayerSSHKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSoftLayerSSHKeyRead,

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"public_key": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"fingerprint": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"notes": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSoftLayerSSHKeyRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetAccountService(sess)

	label := d.Get("name").(string)

	keys, err := service.
		Filter(filter.Build(filter.Path("sshKeys.label").Eq(label))).
		Limit(1).
		GetSshKeys()
	if err != nil {
		return fmt.Errorf("Error retrieving SSH key: %s", err)
	}

	if len(keys) == 0 {
		return fmt.Errorf("No ssh key found with name: %s", label)
	}

	key := keys[0]

	d.SetId(fmt.Sprintf("%d", *key.Id))
	d.Set("name", label)
	d.Set("public_key", strings.TrimSpace(*key.Key))
	d.Set("fingerprint", *key.Fingerprint)

	if keys[0].Notes != nil {
		d.Set("notes", *key.Notes)
	}

	return nil
}
