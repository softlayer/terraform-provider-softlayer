package softlayer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

func resourceSoftLayerNetworkVlan() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerNetworkVlanCreate,
		Read: resourceSoftLayerNetworkVlanRead,
		Update:   resourceSoftLayerNetworkVlanUpdate,
		Delete:   resourceSoftLayerNetworkVlanDelete,
		Exists:   resourceSoftLayerNetworkVlanExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"primary_router_hostname": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceSoftLayerNetworkVlanCreate(d *schema.ResourceData, meta interface{}) error {

	// Need to implement the login to create a vlan
	return resourceSoftLayerNetworkVlanRead(d, meta)
}

func resourceSoftLayerNetworkVlanRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkVlanService(sess)

	vlanId, _ := strconv.Atoi(d.Id())

	vlan, err := service.Id(vlanId).GetObject()
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving Network Vlan: %s", err)
	}

	d.Set("id", *vlan.Id)
	d.Set("name", *vlan.Name)
	d.Set("primary_router_hostname", *vlan.PrimaryRouter)

	return nil
}

func resourceSoftLayerNetworkVlanUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkVlanService(sess)

	vlanId, _ := strconv.Atoi(d.Id())

	opts := datatypes.Network_Vlan{}

	if d.HasChange("name") {
		opts.Name = sl.String(d.Get("name").(string))
	}

	_, err := service.Id(vlanId).EditObject(&opts)

	if err != nil {
		return fmt.Errorf("Error editing Network Vlan: %s", err)
	}
	return nil
}

func resourceSoftLayerNetworkVlanDelete(d *schema.ResourceData, meta interface{}) error {
	// Need to implement logic to delete a vlan
	return nil
}

func resourceSoftLayerNetworkVlanExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(*session.Session)
	service := services.GetNetworkVlanService(sess)

	vlanId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(vlanId).GetObject()
	return result.Id != nil && err == nil && *result.Id == vlanId, nil
}
