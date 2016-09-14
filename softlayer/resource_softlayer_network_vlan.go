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
			"datacenter": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"vlan_number": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"primary_router_hostname": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"price": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"child_resource_count": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"subnets": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"subnet_type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
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

	vlanId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid vlan ID, must be an integer: %s", err)
	}

	vlan, err := service.Id(vlanId).Mask(
		"id," +
			"name," +
			"primaryRouter[datacenter[name]]," +
			"primaryRouter[hostname]," +
			"vlanNumber," +
			"billingItem[recurringFee]," +
			"guestNetworkComponentCount," +
			"subnets[networkIdentifier,cidr,subnetType]",
	).GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving Network Vlan: %s", err)
	}

	d.Set("id", *vlan.Id)
	d.Set("vlan_number", *vlan.VlanNumber)
	d.Set("child_resource_count", *vlan.GuestNetworkComponentCount)
	if vlan.Name != nil {
		d.Set("name", *vlan.Name)
	} else {
		d.Set("name", "")
	}

	if vlan.PrimaryRouter != nil {
		d.Set("primary_router_hostname", *vlan.PrimaryRouter.Hostname)
		if strings.HasPrefix(*vlan.PrimaryRouter.Hostname, "fcr") {
			d.Set("type", "PUBLIC")
		} else {
			d.Set("type", "PRIVATE")
		}
		if vlan.PrimaryRouter.Datacenter != nil {
			d.Set("datacenter", *vlan.PrimaryRouter.Datacenter.Name)
		}
	}

	if vlan.BillingItem != nil {
		d.Set("price", *vlan.BillingItem.RecurringFee)
	} else {
		d.Set("price", 0)
	}

	// Subnets
	subnets := make([]map[string]interface{}, 0)

	for _, elem := range vlan.Subnets {
		subnet := make(map[string]interface{})
		subnet["subnet"] = *elem.NetworkIdentifier + "/" + strconv.Itoa(*elem.Cidr)
		subnet["subnet_type"] = *elem.SubnetType
		subnets = append(subnets, subnet)
	}
	d.Set("subnets", subnets)

	return nil
}

func resourceSoftLayerNetworkVlanUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkVlanService(sess)

	vlanId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid vlan ID, must be an integer: %s", err)
	}

	opts := datatypes.Network_Vlan{}

	if d.HasChange("name") {
		opts.Name = sl.String(d.Get("name").(string))
	}

	_, err = service.Id(vlanId).EditObject(&opts)

	if err != nil {
		return fmt.Errorf("Error updating Network Vlan: %s", err)
	}
	return resourceSoftLayerNetworkVlanRead(d, meta)
}

func resourceSoftLayerNetworkVlanDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkVlanService(sess)

	vlanId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid vlan ID, must be an integer: %s", err)
	}

	// Check if the VLAN is existed.
	_, err = service.Id(vlanId).Mask("id").GetObject()
	if err != nil {
		return fmt.Errorf("Error deleting Network Vlan: %s", err)
	}

	billingItem, err := service.Id(vlanId).GetBillingItem()
	if err != nil {
		return fmt.Errorf("Error deleting Network Vlan: %s", err)
	}
	if billingItem.Id == nil {
		return nil
	}

	success, err := services.GetBillingItemService(sess).Id(*billingItem.Id).CancelService()
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("SoftLayer reported an unsuccessful cancellation")
	}

	return nil
}

func resourceSoftLayerNetworkVlanExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(*session.Session)
	service := services.GetNetworkVlanService(sess)

	vlanId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid vlan ID, must be an integer: %s", err)
	}

	_, err = service.Id(vlanId).Mask("id").GetObject()

	if err != nil {
		return false, err
	}

	return true, nil
}
