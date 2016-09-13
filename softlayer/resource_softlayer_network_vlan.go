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
		Read:     resourceSoftLayerNetworkVlanRead,
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
			"datacenter": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"note": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"primary_router_hostname": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"vlan_number": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
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
	mask := strings.Join([]string{
		"id",
		"name",
		"primaryRouter.datacenter.name",
		"type.name",
		"note",
		"primaryRouter.hostname",
		"vlanNumber",
	}, ";")

	vlan, err := service.Id(vlanId).Mask(mask).GetObject()
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving Network Vlan: %s", err)
	}

	d.Set("id", *vlan.Id)
	d.Set("name", *vlan.Name)
	if vlan.Type != nil {
		d.Set("type", *vlan.Type.Name)
	}
	if vlan.Note != nil {
		d.Set("note", *vlan.Note)
	}
	d.Set("vlan_number", *vlan.VlanNumber)

	if vlan.PrimaryRouter != nil {
		d.Set("primary_router_hostname", *vlan.PrimaryRouter.Hostname)
		if vlan.PrimaryRouter.Datacenter != nil {
			d.Set("datacenter", *vlan.PrimaryRouter.Datacenter.Name)
		}
	}

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
	sess := meta.(*session.Session)
	accountService := services.GetAccountService(sess)
	billingItemService := services.GetBillingItemService(sess)

	vlanId, _ := strconv.Atoi(d.Id())

	mask := strings.Join([]string{
		"id",
		"billingItem.id",
	}, ";")

	filter := fmt.Sprintf(
		`{"networkVlans":{"id":{"operation":"%s"}}`, vlanId,
	)

	networkVlans, err := accountService.Mask(mask).Filter(filter).GetNetworkVlans()
	if err != nil {
		return nil, fmt.Errorf("Error retrieving list of Network Vlans: %s", err)
	}

	if len(networkVlans) < 1 {
		return nil, fmt.Errorf(
			"Unable to locate a vlan matching the provided id: %s", vlanId)
	}

	billingItemId := networkVlans[0].BillingItem.Id
	_, err := billingItemService.Id(billingItemId).CancelItem()

	if err != nil {
		return fmt.Errorf("Error deleting Network Vlan: %s", err)
	}

	d.SetId("")

	return nil
}

func resourceSoftLayerNetworkVlanExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(*session.Session)
	service := services.GetNetworkVlanService(sess)

	vlanId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid Id, Id must be an integer: %s", err)
	}

	result, err := service.Id(vlanId).GetObject()
	return result.Id != nil && err == nil && *result.Id == vlanId, nil
}
