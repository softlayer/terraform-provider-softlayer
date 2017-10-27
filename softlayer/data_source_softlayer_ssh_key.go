package softlayer

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
)

func dataSourceSoftLayerSSHKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSoftLayerSSHKeyRead,

		Schema: map[string]*schema.Schema{
			"label": &schema.Schema{
				Description: "The label associated with the ssh key",
				Type:        schema.TypeString,
				Required:    true,
			},

			"public_key": &schema.Schema{
				Description: "The public ssh key",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"fingerprint": &schema.Schema{
				Description: "A sequence of bytes to authenticate or lookup a longer ssh key",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"notes": &schema.Schema{
				Description: "A small note about a ssh key to use at your discretion",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"most_recent": &schema.Schema{
				Description: "If true and multiple entries are found, the most recently created key is used. " +
					"If false, an error is returned",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func dataSourceSoftLayerSSHKeyRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetAccountService(sess)

	label := d.Get("label").(string)
	most_recent := d.Get("most_recent").(bool)

	keys, err := service.
		Filter(filter.Build(filter.Path("sshKeys.label").Eq(label))).
		Mask("id,label,key,fingerprint,notes,createDate").
		GetSshKeys()

	if err != nil {
		return fmt.Errorf("Error retrieving SSH key: %s", err)
	}

	if len(keys) > 1 && !most_recent {
		return fmt.Errorf(
			"More than one ssh key found with label matching [%s]. "+
				"Either set 'most_recent' to true in your "+
				"configuration to force the most recent ssh key "+
				"to be used, or ensure that the label is unique", label)
	}

	if len(keys) == 0 {
		return fmt.Errorf("No ssh key found with name [%s]", label)
	}

	var key datatypes.Security_Ssh_Key
	if len(keys) > 0 {
		// find key with most recent create date
		key = keys[0]
		for i := 1; i < len(keys); i++ {
			if keys[i].CreateDate.After(key.CreateDate.Time) {
				key = keys[i]
			}
		}
	} else {
		key = keys[0]
	}

	d.SetId(fmt.Sprintf("%d", *key.Id))
	d.Set("name", label)
	d.Set("public_key", strings.TrimSpace(*key.Key))
	d.Set("fingerprint", *key.Fingerprint)

	if keys[0].Notes != nil {
		d.Set("notes", *key.Notes)
	}

	return nil
}
